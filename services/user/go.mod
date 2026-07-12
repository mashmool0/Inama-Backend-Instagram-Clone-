module github.com/mashmool0/inama/services/user

go 1.23

require (
	github.com/mashmool0/inama/libs v0.0.0
	github.com/mashmool0/inama/proto/gen v0.0.0
	google.golang.org/grpc v1.68.0
)

require (
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.36.0 // indirect
)

replace github.com/mashmool0/inama/libs => ../../libs/go

replace github.com/mashmool0/inama/proto/gen => ../../proto/gen
