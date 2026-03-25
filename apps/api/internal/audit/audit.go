package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	ID         uuid.UUID      `json:"id"`
	ActorID    uuid.UUID      `json:"actor_id"`
	EventType  string         `json:"event_type"`
	TargetType string         `json:"target_type"`
	TargetID   uuid.UUID      `json:"target_id"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"created_at"`
}

type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

func (s *Service) Log(ctx context.Context, actorID uuid.UUID, eventType string, targetType string, targetID uuid.UUID, payload map[string]any) error {
	bytes, _ := json.Marshal(payload)
	_, err := s.db.Exec(ctx, `INSERT INTO audit_events(actor_id,event_type,target_type,target_id,payload) VALUES ($1,$2,$3,$4,$5)`, actorID, eventType, targetType, targetID, bytes)
	return err
}

func (s *Service) List(ctx context.Context, limit int) ([]Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.Query(ctx, `SELECT id, actor_id, event_type, target_type, target_id, payload, created_at FROM audit_events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event
		var raw []byte
		if err := rows.Scan(&e.ID, &e.ActorID, &e.EventType, &e.TargetType, &e.TargetID, &raw, &e.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &e.Payload)
		out = append(out, e)
	}
	return out, rows.Err()
}
