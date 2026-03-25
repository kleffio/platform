package ports

import (
	"context"

	"github.com/kleff/platform/internal/core/catalog/domain"
)

type CrateRepository interface {
	ListCrates(ctx context.Context) ([]*domain.Crate, error)
	GetCrate(ctx context.Context, id string) (*domain.Crate, error)
}

type BlueprintRepository interface {
	ListBlueprints(ctx context.Context, crateID string) ([]*domain.Blueprint, error)
	GetBlueprint(ctx context.Context, id string) (*domain.Blueprint, error)
}
