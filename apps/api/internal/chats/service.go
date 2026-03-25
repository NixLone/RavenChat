package chats

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Broadcaster interface {
	BroadcastChat(chatID uuid.UUID, event any)
}

type AuditSink interface {
	Log(ctx context.Context, actorID uuid.UUID, eventType string, targetType string, targetID uuid.UUID, payload map[string]any) error
}

type Service struct {
	repo        *Repository
	broadcaster Broadcaster
	audit       AuditSink
}

func NewService(repo *Repository, broadcaster Broadcaster, audit AuditSink) *Service {
	return &Service{repo: repo, broadcaster: broadcaster, audit: audit}
}

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]Chat, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *Service) Messages(ctx context.Context, chatID uuid.UUID, limit int) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.Messages(ctx, chatID, limit)
}

func (s *Service) SendMessage(ctx context.Context, chatID, senderID uuid.UUID, body string, fileID *uuid.UUID) (Message, error) {
	m, err := s.repo.SaveMessage(ctx, Message{ChatID: chatID, SenderID: senderID, Body: body, Status: "sent", FileID: fileID})
	if err != nil {
		return Message{}, err
	}
	s.broadcaster.BroadcastChat(chatID, map[string]any{"type": "message.created", "message": m})
	return m, nil
}

func (s *Service) EditMessage(ctx context.Context, actorID, messageID uuid.UUID, createdAt time.Time, body string) error {
	if time.Since(createdAt) > 15*time.Minute {
		return errors.New("edit window exceeded")
	}
	if err := s.repo.EditMessage(ctx, messageID, body); err != nil {
		return err
	}
	return s.audit.Log(ctx, actorID, "message.edited", "message", messageID, map[string]any{"edited": true})
}

func (s *Service) DeleteMessage(ctx context.Context, actorID, messageID uuid.UUID) error {
	if err := s.repo.DeleteMessage(ctx, messageID); err != nil {
		return err
	}
	return s.audit.Log(ctx, actorID, "message.deleted", "message", messageID, map[string]any{"deleted": true})
}
