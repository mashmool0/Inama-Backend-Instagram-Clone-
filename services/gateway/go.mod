module github.com/mashmool0/inama/services/gateway

go 1.25.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/mashmool0/inama/libs v0.0.0
	github.com/mashmool0/inama/proto/gen v0.0.0
	github.com/redis/go-redis/v9 v9.7.0
	google.golang.org/grpc v1.82.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/mashmool0/inama/libs => ../../libs/go
	github.com/mashmool0/inama/proto/gen => ../../proto/gen
)
