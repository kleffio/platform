// Package application contains the use-case logic for the notifications module.
package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/kleffio/platform/internal/core/notifications/domain"
	"github.com/kleffio/platform/internal/core/notifications/ports"
	"github.com/kleffio/platform/internal/shared/ids"
)

// Service orchestrates notification operations and delivers real-time events
// to connected SSE clients via the Hub.
type Service struct {
	repo   ports.NotificationRepository
	hub    *Hub
	logger *slog.Logger
}

// NewService creates a Service.
func NewService(repo ports.NotificationRepository, hub *Hub, logger *slog.Logger) *Service {
	return &Service{repo: repo, hub: hub, logger: logger}
}

// CreateInput holds the fields needed to create a new notification.
type CreateInput struct {
	UserID string
	Type   domain.Type
	Title  string
	Body   string
	Data   map[string]any
}

// Create persists a new notification and pushes it to any live SSE connections
// the user currently has open.
func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Notification, error) {
	n := &domain.Notification{
		ID:        ids.New(),
		UserID:    in.UserID,
		Type:      in.Type,
		Title:     in.Title,
		Body:      in.Body,
		Data:      in.Data,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, n); err != nil {
		return nil, err
	}

	// Best-effort push to live SSE connections; never fails the caller.
	s.hub.Push(in.UserID, n)

	return n, nil
}

// List returns notifications for userID according to f.
func (s *Service) List(ctx context.Context, userID string, f domain.ListFilter) ([]*domain.Notification, error) {
	return s.repo.List(ctx, userID, f)
}

// CountUnread returns the number of unread notifications for userID.
func (s *Service) CountUnread(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

// MarkRead marks a single notification as read. The notification must belong to userID.
func (s *Service) MarkRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkRead(ctx, id, userID)
}

// MarkAllRead marks every unread notification for userID as read.
func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}

// MarkReadByInviteID marks the project_invitation notification for this invite as read.
func (s *Service) MarkReadByInviteID(ctx context.Context, userID, inviteID string) error {
	return s.repo.MarkReadByInviteID(ctx, userID, inviteID)
}

// Delete removes a notification. The notification must belong to userID.
func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
