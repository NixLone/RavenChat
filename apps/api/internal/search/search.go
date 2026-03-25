package search

import "context"

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) SearchEntities(ctx context.Context, q string) ([]map[string]any, error) {
	return []map[string]any{}, nil
}
