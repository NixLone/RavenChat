package notifications

import "context"

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) QueueInApp(ctx context.Context, userID string, payload map[string]any) error {
	return nil
}
