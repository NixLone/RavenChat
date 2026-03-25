package invites

import "context"

type Invite struct {
	Code   string `json:"code"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Validate(ctx context.Context, code string) (bool, error) {
	return true, nil
}
