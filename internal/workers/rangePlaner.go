
package workers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/YacineMK/nasa-space-apps/internal/types"
	"github.com/redis/go-redis/v9"
)

func (w *Workers) RangePlanerWorker(
	ctx context.Context,
	startDate time.Time,
	endDate time.Time,
	duration time.Duration,
	latitude float64,
	longitude float64,
) error {
	if duration <= 0 {
		return errors.New("window must be positive")
	}

	if !startDate.Before(endDate) {
		return errors.New("start date must be before end date")
	}

	for current := startDate; current.Before(endDate); {
		next := current.Add(duration)

		if next.After(endDate) {
			next = endDate
		}

		job := types.RangeJob{
			StartDate: current.Format(time.RFC3339),
			EndDate:   next.Format(time.RFC3339),
			Latitude:  latitude,
			Longitude: longitude,
		}

		payload, err := json.Marshal(job)
		if err != nil {
			return err
		}

		_, err = w.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: types.StreamsList[0].Subject,
			Values: map[string]interface{}{
				"payload": string(payload),
			},
		}).Result()
		if err != nil {
			return err
		}

		current = next
	}

	return nil
}