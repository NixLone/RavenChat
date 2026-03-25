package boards

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Column struct {
	ID      uuid.UUID `json:"id"`
	BoardID uuid.UUID `json:"board_id"`
	Name    string    `json:"name"`
	Sort    int       `json:"sort_order"`
}

type Task struct {
	ID          uuid.UUID  `json:"id"`
	BoardID     uuid.UUID  `json:"board_id"`
	ColumnID    uuid.UUID  `json:"column_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	CreatorID   uuid.UUID  `json:"creator_id"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Columns(ctx context.Context, boardID uuid.UUID) ([]Column, error) {
	rows, err := r.db.Query(ctx, `SELECT id, board_id, name, sort_order FROM board_columns WHERE board_id=$1 ORDER BY sort_order`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Column
	for rows.Next() {
		var c Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Sort); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) Tasks(ctx context.Context, boardID uuid.UUID) ([]Task, error) {
	rows, err := r.db.Query(ctx, `SELECT id, board_id, column_id, title, description, assignee_id, creator_id, due_date, status, created_at
	FROM tasks WHERE board_id=$1 ORDER BY created_at DESC`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.BoardID, &t.ColumnID, &t.Title, &t.Description, &t.AssigneeID, &t.CreatorID, &t.DueDate, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) CreateTask(ctx context.Context, t Task) (Task, error) {
	t.ID = uuid.New()
	err := r.db.QueryRow(ctx, `INSERT INTO tasks(id, board_id, column_id, title, description, assignee_id, creator_id, due_date, status)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING created_at`, t.ID, t.BoardID, t.ColumnID, t.Title, t.Description, t.AssigneeID, t.CreatorID, t.DueDate, t.Status).Scan(&t.CreatedAt)
	return t, err
}

func (r *Repository) MoveTask(ctx context.Context, taskID, columnID uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE tasks SET column_id=$2, status=$3 WHERE id=$1`, taskID, columnID, status)
	return err
}
