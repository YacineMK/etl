package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YacineMK/nasa-space-apps/internal/repository"
	"github.com/YacineMK/nasa-space-apps/internal/types"
	weatherhttp "github.com/YacineMK/nasa-space-apps/internal/utils/http"
	"github.com/jackc/pgx/v5/pgtype"
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

				if err := w.persistWeather(ctx, *weather); err != nil {
					log.Printf("[%s] Persist failed for message %s: %v",
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

func (w *Workers) persistWeather(
	ctx context.Context,
	weather types.WeatherResponse,
) error {
	days := len(weather.Daily.Time)
	if len(weather.Daily.Precipitation) != days ||
		len(weather.Daily.TemperatureMax) != days ||
		len(weather.Daily.TemperatureMin) != days ||
		len(weather.Daily.Evapotranspiration) != days {
		return fmt.Errorf("daily arrays length mismatch")
	}

	for i := 0; i < days; i++ {
		date, err := time.Parse("2006-01-02", weather.Daily.Time[i])
		if err != nil {
			return fmt.Errorf("parse day %q: %w", weather.Daily.Time[i], err)
		}

		params := repository.InsertWeatherDailyParams{
			Latitude:                 weather.Latitude,
			Longitude:                weather.Longitude,
			Date:                     pgtype.Date{Time: date, Valid: true},
			Timezone:                 weather.Timezone,
			PrecipitationSum:         nullFloat(weather.Daily.Precipitation[i]),
			Temperature2mMax:         nullFloat(weather.Daily.TemperatureMax[i]),
			Temperature2mMin:         nullFloat(weather.Daily.TemperatureMin[i]),
			Et0FaoEvapotranspiration: nullFloat(weather.Daily.Evapotranspiration[i]),
		}

		if err := w.repo.InsertWeatherDaily(ctx, params); err != nil {
			return fmt.Errorf("insert %s: %w", weather.Daily.Time[i], err)
		}
	}

	return nil
}

func nullFloat(value *float64) pgtype.Float8 {
	if value == nil {
		return pgtype.Float8{}
	}
	return pgtype.Float8{Float64: *value, Valid: true}
}