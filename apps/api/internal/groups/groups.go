package groups

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Group struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	GroupID uuid.UUID `json:"group_id"`
	UserID  uuid.UUID `json:"user_id"`
	Role    string    `json:"role"`
}

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Group, error) {
	rows, err := r.db.Query(ctx, `SELECT g.id, g.name, g.created_at
	FROM groups g JOIN group_members gm ON gm.group_id=g.id
	WHERE gm.user_id=$1 ORDER BY g.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *Repository) Members(ctx context.Context, groupID uuid.UUID) ([]Member, error) {
	rows, err := r.db.Query(ctx, `SELECT group_id, user_id, role FROM group_members WHERE group_id=$1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.GroupID, &m.UserID, &m.Role); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) AddMember(ctx context.Context, groupID, userID uuid.UUID, role string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO group_members(group_id,user_id,role) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, groupID, userID, role)
	return err
}

func (r *Repository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM group_members WHERE group_id=$1 AND user_id=$2`, groupID, userID)
	return err
}
