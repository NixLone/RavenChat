package chats

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Kind string

const (
	KindDirect Kind = "direct"
	KindGroup  Kind = "group"
)

type Chat struct {
	ID        uuid.UUID `json:"id"`
	Kind      Kind      `json:"kind"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        uuid.UUID  `json:"id"`
	ChatID    uuid.UUID  `json:"chat_id"`
	SenderID  uuid.UUID  `json:"sender_id"`
	Body      string     `json:"body"`
	Status    string     `json:"status"`
	FileID    *uuid.UUID `json:"file_id,omitempty"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Chat, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT c.id, c.kind, c.name, c.created_at
	FROM chats c
	JOIN chat_members cm ON cm.chat_id=c.id
	WHERE cm.user_id=$1 ORDER BY c.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Chat
	for rows.Next() {
		var c Chat
		if err := rows.Scan(&c.ID, &c.Kind, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) Messages(ctx context.Context, chatID uuid.UUID, limit int) ([]Message, error) {
	rows, err := r.db.Query(ctx, `SELECT id, chat_id, sender_id, body_ciphertext, status, file_id, edited_at, deleted_at, created_at
	FROM messages WHERE chat_id=$1 ORDER BY created_at DESC LIMIT $2`, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Body, &m.Status, &m.FileID, &m.EditedAt, &m.DeletedAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append([]Message{m}, out...)
	}
	return out, rows.Err()
}

func (r *Repository) SaveMessage(ctx context.Context, m Message) (Message, error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	err := r.db.QueryRow(ctx, `INSERT INTO messages(id,chat_id,sender_id,body_ciphertext,status,file_id)
	VALUES($1,$2,$3,$4,$5,$6)
	RETURNING created_at`, m.ID, m.ChatID, m.SenderID, m.Body, m.Status, m.FileID).Scan(&m.CreatedAt)
	return m, err
}

func (r *Repository) EditMessage(ctx context.Context, messageID uuid.UUID, body string) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET body_ciphertext=$2, edited_at=now() WHERE id=$1`, messageID, body)
	return err
}

func (r *Repository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE messages SET deleted_at=now(), body_ciphertext='[deleted]' WHERE id=$1`, messageID)
	return err
}
