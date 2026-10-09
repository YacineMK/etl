package streams

import "github.com/redis/go-redis/v9"

type Stream struct {
	Rdb *redis.Client
	StreamsList []string
}

func New(streamsList []string) {

}