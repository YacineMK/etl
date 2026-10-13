
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/YacineMK/etl/internal/repository"
	"github.com/YacineMK/etl/internal/types"
	"github.com/YacineMK/etl/internal/utils/cache"
	"github.com/YacineMK/etl/internal/utils/database"
	"github.com/YacineMK/etl/internal/utils/streams"
	"github.com/YacineMK/etl/internal/workers"
)

func main() {
	ctx := context.Background()

	rdb, err := cache.New(ctx, "127.0.0.1:6379")
	if err != nil {
		log.Fatal("Redis connection error: ", err)
	}
	defer rdb.Close()

	log.Println("Redis connected")

	pool, err := database.New(ctx, "postgres://postgres:postgres@localhost:5432/nasa_space?sslmode=disable")
	if err != nil {
		log.Fatal("Postgres connection error: ", err)
	}
	defer pool.Close()

	log.Println("Postgres connected")

	s := streams.New(rdb)
	defer s.Close()

	if err := s.CreateStream(ctx, types.StreamsList); err != nil {
		log.Fatal("Stream creation error: ", err)
	}

	log.Println("Streams created successfully")

	wrk := workers.New(
		rdb,
		"https://archive-api.open-meteo.com/v1/archive",
		repository.New(pool),
	)

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
			7*24*time.Hour,
			36.7538, 
			3.0588,  
		)
		if err != nil {
			log.Println("RangePlanerWorker error:", err)
			return
		}

		log.Println("RangePlanerWorker finished successfully")
	}()

	const apiConsumers = 4
	for i := 0; i < apiConsumers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			name := fmt.Sprintf("api-consumer-%d", id)
			if err := wrk.ApiConsumers(ctx, name); err != nil {
				log.Printf("ApiConsumers[%s] stopped: %v", name, err)
			}
		}(i)
	}

	const pgWriters = 4
	for i := 0; i < pgWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			name := fmt.Sprintf("pg-writer-%d", id)
			if err := wrk.PostgresWriter(ctx, name, 20); err != nil {
				log.Printf("PostgresWriter[%s] stopped: %v", name, err)
			}
		}(i)
	}

	wg.Wait()
}