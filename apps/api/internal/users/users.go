package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Profile struct {
	UserID       uuid.UUID  `json:"user_id"`
	DisplayName  string     `json:"display_name"`
	JobTitle     string     `json:"job_title"`
	AvatarFileID *uuid.UUID `json:"avatar_file_id,omitempty"`
}

type Repository interface {
	ByLogin(ctx context.Context, login string) (User, error)
	ByID(ctx context.Context, userID uuid.UUID) (User, error)
	Profile(ctx context.Context, userID uuid.UUID) (Profile, error)
	List(ctx context.Context) ([]User, error)
}

type PGRepository struct{ db *pgxpool.Pool }

func NewPGRepository(db *pgxpool.Pool) *PGRepository { return &PGRepository{db: db} }

func (r *PGRepository) ByLogin(ctx context.Context, login string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `SELECT id, username, email, phone, role, password_hash, created_at
	FROM users WHERE username=$1 OR email=$1 OR phone=$1`, login).
		Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &u.Role, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (r *PGRepository) ByID(ctx context.Context, userID uuid.UUID) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `SELECT id, username, email, phone, role, password_hash, created_at FROM users WHERE id=$1`, userID).
		Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &u.Role, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (r *PGRepository) Profile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	var p Profile
	err := r.db.QueryRow(ctx, `SELECT user_id, display_name, job_title, avatar_file_id FROM user_profiles WHERE user_id=$1`, userID).
		Scan(&p.UserID, &p.DisplayName, &p.JobTitle, &p.AvatarFileID)
	return p, err
}

func (r *PGRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(ctx, `SELECT id, username, email, phone, role, password_hash, created_at FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Phone, &u.Role, &u.PasswordHash, &u.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
