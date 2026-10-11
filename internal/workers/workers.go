package workers

import "github.com/redis/go-redis/v9"

type Workers struct {
	rdb *redis.Client
	Uri string
}

func New(rdb *redis.Client, uri string) *Workers {
	return &Workers{
		rdb: rdb,
		Uri: uri,
	}
}
