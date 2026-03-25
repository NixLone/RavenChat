package presence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct{ rdb *redis.Client }

func NewService(rdb *redis.Client) *Service { return &Service{rdb: rdb} }

func (s *Service) MarkOnline(ctx context.Context, userID uuid.UUID) error {
	return s.rdb.Set(ctx, fmt.Sprintf("presence:%s", userID), "online", 2*time.Minute).Err()
}
