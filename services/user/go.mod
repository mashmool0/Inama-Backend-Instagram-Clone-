module github.com/mashmool0/inama/services/user

go 1.23

require (
	github.com/jackc/pgx/v5 v5.7.2
	github.com/mashmool0/inama/libs v0.0.0
	github.com/mashmool0/inama/proto/gen v0.0.0
	github.com/prometheus/client_golang v1.20.5
	google.golang.org/grpc v1.68.0
)

replace github.com/mashmool0/inama/libs => ../../libs/go

replace github.com/mashmool0/inama/proto/gen => ../../proto/gen
