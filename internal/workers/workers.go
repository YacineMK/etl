package workers

import (
	"github.com/YacineMK/etl/internal/repository"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

const requestsPerSecond = 8

type Workers struct {
	rdb     *redis.Client
	Uri     string
	repo    *repository.Queries
	limiter *rate.Limiter
}

func New(rdb *redis.Client, uri string, repo *repository.Queries) *Workers {
	return &Workers{
		rdb:     rdb,
		Uri:     uri,
		repo:    repo,
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond),
	}
}
