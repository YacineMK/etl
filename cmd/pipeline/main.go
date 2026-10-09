package main

import (
	"context"
	"log"

	"github.com/YacineMK/nasa-space-apps/internal/types"
	"github.com/YacineMK/nasa-space-apps/internal/utils/cache"
	"github.com/YacineMK/nasa-space-apps/internal/utils/streams"
)

func main() {
	ctx := context.Background()

	rdb, err := cache.New(ctx, "127.0.0.1:6379")
	if err != nil {
		log.Println("error:", err)
	}
	log.Println("redis connected")

	streamsList := []types.StreamConfig{
		{
			Subject:       "etl:jobs",
			ConsumerGroup: "etl-processors",
		},
		{
			Subject:       "etl:write-jobs",
			ConsumerGroup: "postgres-writers",
		},
	}
	
	s := streams.New(rdb)
	defer s.Close()
	err = s.CreateStream(ctx, streamsList)
	if err != nil {
		panic(err)
	}
	log.Println("All streams have been created successfully")
}
