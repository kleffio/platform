package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kleffio/platform/internal/core/nodes/domain"
	"github.com/kleffio/platform/internal/core/nodes/ports"
)

type PostgresNodeStore struct {
	db *sql.DB
}

func NewPostgresNodeStore(db *sql.DB) ports.NodeRepository {
	return &PostgresNodeStore{db: db}
}

func (s *PostgresNodeStore) FindByID(ctx context.Context, id string) (*domain.Node, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, hostname, region, ip_address, status,
		       total_vcpu, total_mem_gb, total_disk_gb,
		       used_vcpu, used_mem_gb, used_disk_gb,
		       token_hash, last_heartbeat_at, created_at, updated_at
		FROM nodes WHERE id = $1`, id)
	return scanNode(row)
}

func (s *PostgresNodeStore) FindByHostname(ctx context.Context, hostname string) (*domain.Node, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, hostname, region, ip_address, status,
		       total_vcpu, total_mem_gb, total_disk_gb,
		       used_vcpu, used_mem_gb, used_disk_gb,
		       token_hash, last_heartbeat_at, created_at, updated_at
		FROM nodes WHERE hostname = $1`, hostname)
	return scanNode(row)
}

func (s *PostgresNodeStore) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.Node, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, hostname, region, ip_address, status,
		       total_vcpu, total_mem_gb, total_disk_gb,
		       used_vcpu, used_mem_gb, used_disk_gb,
		       token_hash, last_heartbeat_at, created_at, updated_at
		FROM nodes WHERE token_hash = $1`, tokenHash)
	return scanNode(row)
}

func (s *PostgresNodeStore) ListByRegion(ctx context.Context, region string) ([]*domain.Node, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, hostname, region, ip_address, status,
		       total_vcpu, total_mem_gb, total_disk_gb,
		       used_vcpu, used_mem_gb, used_disk_gb,
		       token_hash, last_heartbeat_at, created_at, updated_at
		FROM nodes WHERE region = $1 ORDER BY created_at DESC`, region)
	if err != nil {
		return nil, fmt.Errorf("list nodes by region: %w", err)
	}
	defer rows.Close()
	return scanNodes(rows)
}

func (s *PostgresNodeStore) ListAll(ctx context.Context) ([]*domain.Node, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, hostname, region, ip_address, status,
		       total_vcpu, total_mem_gb, total_disk_gb,
		       used_vcpu, used_mem_gb, used_disk_gb,
		       token_hash, last_heartbeat_at, created_at, updated_at
		FROM nodes ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	defer rows.Close()
	return scanNodes(rows)
}

func (s *PostgresNodeStore) Save(ctx context.Context, node *domain.Node) error {
	if node == nil {
		return fmt.Errorf("node is required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO nodes (
			id, hostname, region, ip_address, status,
			total_vcpu, total_mem_gb, total_disk_gb,
			used_vcpu, used_mem_gb, used_disk_gb,
			token_hash, last_heartbeat_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15
		)
		ON CONFLICT (id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			region = EXCLUDED.region,
			ip_address = EXCLUDED.ip_address,
			status = EXCLUDED.status,
			total_vcpu = EXCLUDED.total_vcpu,
			total_mem_gb = EXCLUDED.total_mem_gb,
			total_disk_gb = EXCLUDED.total_disk_gb,
			used_vcpu = EXCLUDED.used_vcpu,
			used_mem_gb = EXCLUDED.used_mem_gb,
			used_disk_gb = EXCLUDED.used_disk_gb,
			token_hash = EXCLUDED.token_hash,
			last_heartbeat_at = EXCLUDED.last_heartbeat_at,
			updated_at = EXCLUDED.updated_at`,
		node.ID,
		node.Hostname,
		node.Region,
		node.IPAddress,
		node.Status,
		node.TotalVCPU,
		node.TotalMemGB,
		node.TotalDiskGB,
		node.UsedVCPU,
		node.UsedMemGB,
		node.UsedDiskGB,
		node.TokenHash,
		node.LastHeartbeatAt,
		node.CreatedAt,
		node.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save node: %w", err)
	}
	return nil
}

type nodeScanner interface {
	Scan(dest ...any) error
}

func scanNode(s nodeScanner) (*domain.Node, error) {
	var n domain.Node
	var heartbeat sql.NullTime
	err := s.Scan(
		&n.ID,
		&n.Hostname,
		&n.Region,
		&n.IPAddress,
		&n.Status,
		&n.TotalVCPU,
		&n.TotalMemGB,
		&n.TotalDiskGB,
		&n.UsedVCPU,
		&n.UsedMemGB,
		&n.UsedDiskGB,
		&n.TokenHash,
		&heartbeat,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if heartbeat.Valid {
		n.LastHeartbeatAt = heartbeat.Time.UTC()
	}
	n.CreatedAt = n.CreatedAt.UTC()
	n.UpdatedAt = n.UpdatedAt.UTC()
	return &n, nil
}

func scanNodes(rows *sql.Rows) ([]*domain.Node, error) {
	var out []*domain.Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
