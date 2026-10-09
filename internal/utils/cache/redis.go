package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context,cacheUri string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cacheUri,
		Password: "",
		DB: 0,
	})
	
	_ , err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil , err
	}
	return rdb,nil
}



