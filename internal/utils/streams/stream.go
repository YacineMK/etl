package streams

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YacineMK/etl/internal/types"
	"github.com/redis/go-redis/v9"
)

type Stream struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Stream {
	return &Stream{
		rdb: rdb,
	}
}

func (s *Stream) CreateStream(ctx context.Context, streamsList []types.StreamConfig) error {
	for _, stream := range streamsList {
		err := s.rdb.XGroupCreateMkStream(
			ctx,
			stream.Subject,
			stream.ConsumerGroup,
			"0",
		).Err()
		if err != nil {
			if strings.Contains(err.Error(), "BUSYGROUP") {
				continue
			}
			return fmt.Errorf(
				"create stream %q and group %q: %w",
				stream.Subject,
				stream.ConsumerGroup,
				err,
			)
		}
		log.Printf(
			"Stream %q and consumer group %q initialized",
			stream.Subject,
			stream.ConsumerGroup,
		)
	}
	return nil
}


func (s *Stream) Close() {
	s.rdb.Close()
}