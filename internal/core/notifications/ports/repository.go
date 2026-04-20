// Package ports defines the outbound interfaces for the notifications module.
package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/notifications/domain"
)

// NotificationRepository is the persistence contract for notifications.
type NotificationRepository interface {
	// Save persists a new notification.
	Save(ctx context.Context, n *domain.Notification) error

	// FindByID returns a single notification. Returns sql.ErrNoRows when not found.
	FindByID(ctx context.Context, id string) (*domain.Notification, error)

	// List returns notifications for a user according to the supplied filter.
	List(ctx context.Context, userID string, f domain.ListFilter) ([]*domain.Notification, error)

	// CountUnread returns the number of unread notifications for a user.
	CountUnread(ctx context.Context, userID string) (int, error)

	// MarkRead sets read_at to now for a single notification owned by userID.
	MarkRead(ctx context.Context, id, userID string) error

	// MarkAllRead sets read_at to now for every unread notification owned by userID.
	MarkAllRead(ctx context.Context, userID string) error

	// MarkReadByInviteID marks as read any unread project_invitation notification for userID
	// whose data->>'invite_id' matches inviteID.
	MarkReadByInviteID(ctx context.Context, userID, inviteID string) error

	// Delete removes a notification owned by userID.
	Delete(ctx context.Context, id, userID string) error
}
