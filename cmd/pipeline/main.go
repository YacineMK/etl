
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/YacineMK/nasa-space-apps/internal/types"
	"github.com/YacineMK/nasa-space-apps/internal/utils/cache"
	"github.com/YacineMK/nasa-space-apps/internal/utils/streams"
	"github.com/YacineMK/nasa-space-apps/internal/workers"
)

func main() {
	ctx := context.Background()

	rdb, err := cache.New(ctx, "127.0.0.1:6379")
	if err != nil {
		log.Fatal("Redis connection error: ", err)
	}
	defer rdb.Close()

	log.Println("Redis connected")

	s := streams.New(rdb)
	defer s.Close()

	if err := s.CreateStream(ctx, types.StreamsList); err != nil {
		log.Fatal("Stream creation error: ", err)
	}

	log.Println("Streams created successfully")

	wrk := workers.New(rdb)

	startDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := wrk.RangePlanerWorker(
			ctx,
			startDate,
			endDate,
			24*time.Hour,
			36.7538, 
			3.0588,  
		)
		if err != nil {
			log.Println("RangePlanerWorker error:", err)
			return
		}

		log.Println("RangePlanerWorker finished successfully")
	}()

	wg.Wait()
}