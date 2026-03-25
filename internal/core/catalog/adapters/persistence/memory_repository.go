package persistence

import (
	"context"
	"fmt"
	"sync"

	"github.com/kleff/platform/internal/core/catalog/domain"
)

// MemoryRepository is an in-memory implementation of the catalog repositories.
// It is seeded at startup from YAML files and is read-only at runtime.
type MemoryRepository struct {
	mu         sync.RWMutex
	crates     map[string]*domain.Crate
	blueprints map[string]*domain.Blueprint
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		crates:     make(map[string]*domain.Crate),
		blueprints: make(map[string]*domain.Blueprint),
	}
}

func (r *MemoryRepository) AddCrate(c *domain.Crate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.crates[c.ID] = c
}

func (r *MemoryRepository) AddBlueprint(b *domain.Blueprint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blueprints[b.ID] = b
}

func (r *MemoryRepository) ListCrates(_ context.Context) ([]*domain.Crate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*domain.Crate, 0, len(r.crates))
	for _, c := range r.crates {
		out = append(out, c)
	}
	return out, nil
}

func (r *MemoryRepository) GetCrate(_ context.Context, id string) (*domain.Crate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.crates[id]
	if !ok {
		return nil, fmt.Errorf("crate not found: %s", id)
	}
	return c, nil
}

func (r *MemoryRepository) ListBlueprints(_ context.Context, crateID string) ([]*domain.Blueprint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Blueprint
	for _, b := range r.blueprints {
		if b.CrateID == crateID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (r *MemoryRepository) GetBlueprint(_ context.Context, id string) (*domain.Blueprint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.blueprints[id]
	if !ok {
		return nil, fmt.Errorf("blueprint not found: %s", id)
	}
	return b, nil
}
