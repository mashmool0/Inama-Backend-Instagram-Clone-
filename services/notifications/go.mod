module github.com/mashmool0/inama/services/notifications

go 1.23

require (
	github.com/jackc/pgx/v5 v5.7.2
	github.com/mashmool0/inama/libs v0.0.0
	github.com/mashmool0/inama/proto/gen v0.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
	google.golang.org/grpc v1.68.0
)

replace github.com/mashmool0/inama/libs => ../../libs/go

replace github.com/mashmool0/inama/proto/gen => ../../proto/gen
