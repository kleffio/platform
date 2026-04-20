// Package persistence provides the PostgreSQL implementation of NotificationRepository.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/notifications/domain"
	"github.com/kleffio/platform/internal/core/notifications/ports"
)

// PostgresNotificationStore implements ports.NotificationRepository.
type PostgresNotificationStore struct {
	db *sql.DB
}

// NewPostgresNotificationStore returns a store backed by db.
func NewPostgresNotificationStore(db *sql.DB) ports.NotificationRepository {
	return &PostgresNotificationStore{db: db}
}

// ── Write operations ──────────────────────────────────────────────────────────

func (s *PostgresNotificationStore) Save(ctx context.Context, n *domain.Notification) error {
	data, err := json.Marshal(n.Data)
	if err != nil {
		return fmt.Errorf("marshal notification data: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		n.ID, n.UserID, string(n.Type), n.Title, n.Body, data, n.CreatedAt)
	if err != nil {
		return fmt.Errorf("save notification: %w", err)
	}
	return nil
}

func (s *PostgresNotificationStore) MarkRead(ctx context.Context, id, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE notifications SET read_at = $1
		WHERE id = $2 AND user_id = $3 AND read_at IS NULL`,
		time.Now().UTC(), id, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

func (s *PostgresNotificationStore) MarkAllRead(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE notifications SET read_at = $1
		WHERE user_id = $2 AND read_at IS NULL`,
		time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

func (s *PostgresNotificationStore) Delete(ctx context.Context, id, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM notifications WHERE id = $1 AND user_id = $2`,
		id, userID)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

// ── Read operations ───────────────────────────────────────────────────────────

func (s *PostgresNotificationStore) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, type, title, body, data, read_at, created_at
		FROM notifications WHERE id = $1`, id)
	return scanNotification(row)
}

func (s *PostgresNotificationStore) List(ctx context.Context, userID string, f domain.ListFilter) ([]*domain.Notification, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	var rows *sql.Rows
	var err error

	if f.UnreadOnly {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, type, title, body, data, read_at, created_at
			FROM notifications
			WHERE user_id = $1 AND read_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`,
			userID, limit, offset)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, type, title, body, data, read_at, created_at
			FROM notifications
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`,
			userID, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	return scanNotifications(rows)
}

func (s *PostgresNotificationStore) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`,
		userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return count, nil
}

// ── Scanners ──────────────────────────────────────────────────────────────────

func scanNotification(row *sql.Row) (*domain.Notification, error) {
	var n domain.Notification
	var readAt sql.NullTime
	var rawData []byte
	if err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &rawData, &readAt, &n.CreatedAt); err != nil {
		return nil, err
	}
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}
	if len(rawData) > 0 {
		_ = json.Unmarshal(rawData, &n.Data)
	}
	return &n, nil
}

func scanNotifications(rows *sql.Rows) ([]*domain.Notification, error) {
	var out []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		var readAt sql.NullTime
		var rawData []byte
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &rawData, &readAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}
		if len(rawData) > 0 {
			_ = json.Unmarshal(rawData, &n.Data)
		}
		out = append(out, &n)
	}
	return out, rows.Err()
}
