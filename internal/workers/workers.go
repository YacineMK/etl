package workers

import (
	"github.com/YacineMK/etl/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Workers struct {
	rdb  *redis.Client
	Uri  string
	repo *repository.Queries
}

func New(rdb *redis.Client, uri string, repo *repository.Queries) *Workers {
	return &Workers{
		rdb:  rdb,
		Uri:  uri,
		repo: repo,
	}
}
