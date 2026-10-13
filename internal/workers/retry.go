package workers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/YacineMK/etl/internal/types"
	"github.com/redis/go-redis/v9"
)

const (
	maxAttempts = 3
	retryDelay  = 5 * time.Second
)

func (w *Workers) RetryWorker(ctx context.Context) error {
	stream := types.StreamsList[2]

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		result, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    stream.ConsumerGroup,
			Consumer: "retry-worker",
			Streams:  []string{stream.Subject, ">"},
			Count:    20,
			Block:    3 * time.Second,
		}).Result()

		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("[retry] Redis read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		time.Sleep(retryDelay)

		for _, resultStream := range result {
			for _, msg := range resultStream.Messages {
				payload, ok := msg.Values["payload"].(string)
				if !ok {
					log.Printf("[retry] Message %s has invalid payload", msg.ID)
					w.ack(ctx, stream, msg.ID)
					continue
				}

				if _, err := w.rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: types.StreamsList[0].Subject,
					Values: map[string]interface{}{
						"payload": payload,
					},
				}).Result(); err != nil {
					log.Printf("[retry] Requeue message %s failed: %v", msg.ID, err)
					continue
				}

				w.ack(ctx, stream, msg.ID)
			}
		}
	}
}

// requeue sends a failed job back to the retry stream, or to the dead-letter
// stream once it has exhausted its attempts, then acknowledges the original.
func (w *Workers) requeue(
	ctx context.Context,
	stream types.StreamConfig,
	msgID string,
	job types.RangeJob,
	reason error,
) {
	target := types.StreamsList[3].Subject
	if job.Attempt < maxAttempts {
		job.Attempt++
		target = types.StreamsList[2].Subject
	}

	payload, err := json.Marshal(job)
	if err != nil {
		log.Printf("Marshal job for requeue failed: %v", err)
		return
	}

	if _, err := w.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: target,
		Values: map[string]interface{}{
			"payload": string(payload),
			"reason":  reason.Error(),
		},
	}).Result(); err != nil {
		log.Printf("Requeue job %s failed: %v", msgID, err)
		return
	}

	w.ack(ctx, stream, msgID)
}

// deadLetter parks an unparseable message on the dead-letter stream.
func (w *Workers) deadLetter(
	ctx context.Context,
	stream types.StreamConfig,
	msgID string,
	payload string,
	reason string,
) {
	if _, err := w.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: types.StreamsList[3].Subject,
		Values: map[string]interface{}{
			"payload": payload,
			"reason":  reason,
		},
	}).Result(); err != nil {
		log.Printf("Dead-letter message %s failed: %v", msgID, err)
		return
	}

	w.ack(ctx, stream, msgID)
}

func (w *Workers) ack(ctx context.Context, stream types.StreamConfig, msgID string) {
	if err := w.rdb.XAck(
		ctx,
		stream.Subject,
		stream.ConsumerGroup,
		msgID,
	).Err(); err != nil {
		log.Printf("Acknowledge message %s failed: %v", msgID, err)
	}
}
