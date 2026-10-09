package main

import (
	"context"
	"log"

	"github.com/YacineMK/nasa-space-apps/internal/utils/cache"
)

func main() {
	ctx := context.Background()

	_, err := cache.New(ctx, "127.0.0.1:6379")
	if err != nil {
		log.Println("error:", err)
	}
	log.Println("redis connected")

}
