package auth

import (
	"context"
	"errors"

	"ravenchat/apps/api/internal/users"
)

type UserRepository interface {
	ByLogin(ctx context.Context, login string) (users.User, error)
}

type LocalProvider struct {
	repo UserRepository
}

func NewLocalProvider(repo UserRepository) *LocalProvider {
	return &LocalProvider{repo: repo}
}

func (p *LocalProvider) Authenticate(ctx context.Context, login, password string) (Identity, error) {
	u, err := p.repo.ByLogin(ctx, login)
	if err != nil {
		return Identity{}, errors.New("invalid credentials")
	}
	if err := VerifyPassword(u.PasswordHash, password); err != nil {
		return Identity{}, errors.New("invalid credentials")
	}
	return Identity{UserID: u.ID, Username: u.Username, Role: Role(u.Role)}, nil
}
