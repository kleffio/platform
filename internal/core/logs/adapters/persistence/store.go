package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kleffio/platform/internal/core/logs/domain"
)

type PostgresLogStore struct {
	db *sql.DB
}

func NewPostgresLogStore(db *sql.DB) *PostgresLogStore {
	return &PostgresLogStore{db: db}
}

func (s *PostgresLogStore) SaveBatch(ctx context.Context, lines []*domain.LogLine) error {
	if len(lines) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO workload_log_lines (workload_id, project_id, ts, stream, line)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("prepare log insert: %w", err)
	}
	defer stmt.Close()

	for _, l := range lines {
		if _, err := stmt.ExecContext(ctx, l.WorkloadID, l.ProjectID, l.Ts, l.Stream, l.Line); err != nil {
			return fmt.Errorf("insert log line: %w", err)
		}
	}
	return tx.Commit()
}

func (s *PostgresLogStore) ListByWorkload(ctx context.Context, workloadID string, limit int) ([]*domain.LogLine, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, workload_id, project_id, ts, stream, line
		FROM workload_log_lines
		WHERE workload_id = $1
		ORDER BY ts DESC
		LIMIT $2
	`, workloadID, limit)
	if err != nil {
		return nil, fmt.Errorf("list log lines: %w", err)
	}
	defer rows.Close()

	var results []*domain.LogLine
	for rows.Next() {
		l := &domain.LogLine{}
		if err := rows.Scan(&l.ID, &l.WorkloadID, &l.ProjectID, &l.Ts, &l.Stream, &l.Line); err != nil {
			return nil, fmt.Errorf("scan log line: %w", err)
		}
		results = append(results, l)
	}
	// Reverse so results are in chronological order (oldest first).
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results, rows.Err()
}
