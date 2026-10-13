package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YacineMK/etl/internal/repository"
	"github.com/YacineMK/etl/internal/types"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

func (w *Workers) PostgresWriter(
	ctx context.Context,
	consumerName string,
	batchSize int,
) error {
	if batchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}

	stream := types.StreamsList[1]

	records := make([]repository.CopyWeatherDailyParams, 0, batchSize)
	ids := make([]string, 0, batchSize)

	flush := func() error {
		if len(records) == 0 {
			return nil
		}

		written, err := w.repo.CopyWeatherDaily(ctx, records)
		if err != nil {
			return fmt.Errorf("copy %d rows: %w", len(records), err)
		}

		if err := w.rdb.XAck(
			ctx,
			stream.Subject,
			stream.ConsumerGroup,
			ids...,
		).Err(); err != nil {
			return fmt.Errorf("ack %d messages: %w", len(ids), err)
		}

		log.Printf("[%s] Wrote %d rows", consumerName, written)

		records = records[:0]
		ids = ids[:0]
		return nil
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		result, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    stream.ConsumerGroup,
			Consumer: consumerName,
			Streams:  []string{stream.Subject, ">"},
			Count:    int64(batchSize),
			Block:    3 * time.Second,
		}).Result()

		if errors.Is(err, redis.Nil) {
			if err := flush(); err != nil {
				log.Printf("[%s] Flush error: %v", consumerName, err)
			}
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
					log.Printf("[%s] Message %s has invalid payload",
						consumerName, msg.ID)
					continue
				}

				var record types.WeatherRecord
				if err := json.Unmarshal([]byte(payload), &record); err != nil {
					log.Printf("[%s] Decode message %s: %v",
						consumerName, msg.ID, err)
					continue
				}

				params, err := toCopyParams(record)
				if err != nil {
					log.Printf("[%s] Convert message %s: %v",
						consumerName, msg.ID, err)
					continue
				}

				records = append(records, params)
				ids = append(ids, msg.ID)

				if len(records) >= batchSize {
					if err := flush(); err != nil {
						log.Printf("[%s] Flush error: %v", consumerName, err)
					}
				}
			}
		}
	}
}

func toCopyParams(record types.WeatherRecord) (repository.CopyWeatherDailyParams, error) {
	date, err := time.Parse("2006-01-02", record.Date)
	if err != nil {
		return repository.CopyWeatherDailyParams{}, fmt.Errorf("parse date %q: %w", record.Date, err)
	}

	return repository.CopyWeatherDailyParams{
		Latitude:                 record.Latitude,
		Longitude:                record.Longitude,
		Date:                     pgtype.Date{Time: date, Valid: true},
		Timezone:                 record.Timezone,
		PrecipitationSum:         nullFloat(record.PrecipitationSum),
		Temperature2mMax:         nullFloat(record.Temperature2mMax),
		Temperature2mMin:         nullFloat(record.Temperature2mMin),
		Et0FaoEvapotranspiration: nullFloat(record.Et0FaoEvapotranspiration),
	}, nil
}

func nullFloat(value *float64) pgtype.Float8 {
	if value == nil {
		return pgtype.Float8{}
	}
	return pgtype.Float8{Float64: *value, Valid: true}
}
