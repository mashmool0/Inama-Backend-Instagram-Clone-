# Multi-stage build — small final image, no Go toolchain shipped.
FROM golang:1.23-alpine AS build
WORKDIR /src

COPY services/notifications/go.mod services/notifications/go.sum ./services/notifications/
COPY proto/gen/go.mod proto/gen/go.sum ./proto/gen/
COPY libs/go/go.mod ./libs/go/

WORKDIR /src/services/notifications
RUN go mod download

WORKDIR /src
COPY services/notifications ./services/notifications
COPY proto/gen ./proto/gen
COPY libs/go ./libs/go

WORKDIR /src/services/notifications
RUN CGO_ENABLED=0 go build -o /server .

FROM alpine:3.20
RUN apk add --no-cache curl
COPY --from=build /server /server
EXPOSE 8080 50051
CMD ["/server"]
