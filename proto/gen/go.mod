// Generated gRPC/protobuf code for all services lives here as ONE shared Go
// module. Each Go service imports it, e.g.:
//
//   import postspb "github.com/mashmool0/inama/proto/gen/posts"
//
// and in the service's own go.mod:
//
//   require github.com/mashmool0/inama/proto/gen v0.0.0
//   replace github.com/mashmool0/inama/proto/gen => ../../proto/gen
//
// `make gen` runs `go mod tidy` here after generation to fill in the
// grpc/protobuf dependency versions below.
module github.com/mashmool0/inama/proto/gen

go 1.23

require (
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.36.0
)

require (
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)
