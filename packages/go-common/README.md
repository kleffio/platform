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
├── domain/
│   ├── errors.go       # AppError, constructors (NewNotFound, NewUnauthorized…), sentinel errors
│   ├── response.go     # APIResponse[T], PaginatedResponse[T], APIErrorResponse
│   └── pagination.go   # PageRequest, PaginationMeta, offset/meta helpers
├── adapters/http/
│   ├── middleware.go   # RequestID, Logger, CORS, Recover middleware
│   └── response.go     # JSON, Success, Created, NoContent, Paginated, Error helpers
└── bootstrap/
    └── server.go       # RunServer — graceful start/shutdown with signal handling
```

## Usage

### Local development (`go.work` or `replace`)

```
// go.mod in a service
replace github.com/kleff/go-common => ../../packages/go-common
```

Or use a `go.work` file at the repo root:

```
go 1.23
use (
    ./services/billing-service
    ./services/identity-service
    ./services/gameserver-service
    ./packages/go-common
)
```

### Standard error responses

```go
import (
    "github.com/kleff/go-common/domain"
    httputil "github.com/kleff/go-common/adapters/http"
)

func getServer(w http.ResponseWriter, r *http.Request) {
    server, err := repo.FindByID(r.PathValue("id"))
    if err != nil {
        httputil.Error(w, domain.NewNotFound("game server"))
        return
    }
    httputil.Success(w, server)
}
```

### Middleware chain

```go
import (
    httputil "github.com/kleff/go-common/adapters/http"
    "github.com/kleff/go-common/bootstrap"
)

mux := http.NewServeMux()
// ... register routes ...

handler := httputil.RequestID(
    httputil.Logger(logger)(
        httputil.Recover(logger)(
            httputil.CORS("https://app.kleff.io")(mux),
        ),
    ),
)

bootstrap.RunServer(bootstrap.ServerConfig{
    Port:    8080,
    Handler: handler,
    Logger:  logger,
})
```

### Pagination

```go
import "github.com/kleff/go-common/domain"

req := domain.PageRequest{Page: 1, Limit: 20}
req.Normalise()

rows, total := repo.List(ctx, req.Offset(), req.Limit)
meta := req.BuildMeta(total)

httputil.Paginated(w, rows, meta)
```
