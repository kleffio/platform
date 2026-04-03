package registry

import "github.com/kleffio/platform/internal/core/plugins/domain"

// builtinCatalog is empty — all plugins are sourced from the remote registry.
var builtinCatalog = []*domain.CatalogManifest{}
