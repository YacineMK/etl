package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YacineMK/nasa-space-apps/internal/types"
	weatherhttp "github.com/YacineMK/nasa-space-apps/internal/utils/http"
	"github.com/redis/go-redis/v9"
)

func (w *Workers) ApiConsumers(
	ctx context.Context,
	consumerName string,
) error {
	stream := types.StreamsList[0]

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		result, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    stream.ConsumerGroup,
			Consumer: consumerName,
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
			log.Printf("[%s] Redis read error: %v", consumerName, err)
			time.Sleep(time.Second)
			continue
		}

		for _, resultStream := range result {
			for _, msg := range resultStream.Messages {
				payload, ok := msg.Values["payload"].(string)
				if !ok {
					log.Printf("Message %s has invalid payload", msg.ID)
					continue
				}

				var job types.RangeJob
				if err := json.Unmarshal([]byte(payload), &job); err != nil {
					log.Printf("Decode message %s: %v", msg.ID, err)
					continue
				}

				weather, err := weatherhttp.FetchWeather(ctx, w.Uri, job)
				if err != nil {
					log.Printf("[%s] Fetch failed for message %s: %v",
						consumerName, msg.ID, err)
					continue
				}

				fmt.Printf(
					"[%s] Fetched weather for (%f, %f): %d days\n",
					consumerName,
					weather.Latitude,
					weather.Longitude,
					len(weather.Daily.Time),
				)

				if err := w.rdb.XAck(
					ctx,
					stream.Subject,
					stream.ConsumerGroup,
					msg.ID,
				).Err(); err != nil {
					log.Printf("Acknowledge message %s failed: %v",
						msg.ID, err)
					continue
				}

				log.Printf("[%s] Processed message %s",
					consumerName, msg.ID)
			}
		}
	}
}