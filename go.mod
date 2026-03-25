module github.com/kleff/platform

go 1.23

require (
	github.com/google/uuid v1.6.0
	github.com/kleff/go-common v0.0.0
	github.com/redis/go-redis/v9 v9.18.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	go.uber.org/atomic v1.11.0 // indirect
)

replace github.com/kleff/go-common => ./packages
