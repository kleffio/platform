# go-common

Shared Go utilities consumed by all Kleff backend services.

## Module

```
github.com/kleff/go-common
```

Go version: `1.23`

## Structure

```
go-common/
├── adapters/       # Reusable adapter implementations (e.g. HTTP middleware, DB helpers)
├── application/    # Shared application-layer utilities (e.g. request context, pagination)
├── bootstrap/      # Common startup helpers (e.g. config loading, signal handling)
├── domain/         # Shared domain primitives (e.g. errors, value objects)
└── ports/          # Common port interfaces shared across services
```

## Usage

Import in a service's `go.mod`:

```
require github.com/kleff/go-common v0.0.0
```

If developing locally with a monorepo workspace, add a `replace` directive or use a `go.work` file at the repo root:

```
replace github.com/kleff/go-common => ../../packages/go-common
```
