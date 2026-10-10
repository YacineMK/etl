package workers

import "github.com/redis/go-redis/v9"

type Workers struct  {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Workers {
	return &Workers{
		rdb: rdb,
	}
}