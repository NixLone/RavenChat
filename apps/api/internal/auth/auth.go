package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleUser            Role = "user"
	RoleSystemAdmin     Role = "system_admin"
	RoleSecurityAuditor Role = "security_auditor"
)

type Identity struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     Role      `json:"role"`
}

type Provider interface {
	Authenticate(ctx context.Context, login, password string) (Identity, error)
}

type JWTManager struct {
	secret []byte
}

type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"un"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{secret: []byte(secret)}
}

func (j *JWTManager) Mint(identity Identity) (string, error) {
	claims := Claims{
		UserID:   identity.UserID.String(),
		Username: identity.Username,
		Role:     string(identity.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.secret)
}

func (j *JWTManager) Parse(token string) (Identity, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil || !parsed.Valid {
		return Identity{}, errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return Identity{}, errors.New("invalid claims")
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return Identity{}, errors.New("invalid user id in token")
	}
	return Identity{UserID: uid, Username: claims.Username, Role: Role(claims.Role)}, nil
}

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func VerifyPassword(hash, p string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
}
