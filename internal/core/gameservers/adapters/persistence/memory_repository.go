package persistence

import (
	"context"
	"fmt"
	"sync"

	"github.com/kleff/platform/internal/core/gameservers/domain"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	servers map[string]*domain.GameServer
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{servers: make(map[string]*domain.GameServer)}
}

func (r *MemoryRepository) Save(_ context.Context, gs *domain.GameServer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servers[gs.ID] = gs
	return nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (*domain.GameServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	gs, ok := r.servers[id]
	if !ok {
		return nil, fmt.Errorf("game server not found: %s", id)
	}
	return gs, nil
}

func (r *MemoryRepository) ListByOrg(_ context.Context, orgID string) ([]*domain.GameServer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.GameServer
	for _, gs := range r.servers {
		if gs.OrganizationID == orgID {
			out = append(out, gs)
		}
	}
	return out, nil
}

func (r *MemoryRepository) UpdateStatus(_ context.Context, id string, status domain.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	gs, ok := r.servers[id]
	if !ok {
		return fmt.Errorf("game server not found: %s", id)
	}
	gs.Status = status
	return nil
}
