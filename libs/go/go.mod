// Shared Go conventions used by every Go service (gateway, user, posts, feed,
// notifications). Consumed like proto/gen:
//
//   require github.com/mashmool0/inama/libs v0.0.0
//   replace github.com/mashmool0/inama/libs => ../../libs/go
//
// `logging` and `config` are pure stdlib. `identity` needs grpc — its
// dependency is resolved by `go mod tidy` (run automatically the first time
// you build a service that imports it).
module github.com/mashmool0/inama/libs

go 1.23

require google.golang.org/grpc v1.68.0
