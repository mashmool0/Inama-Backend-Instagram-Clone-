# Multi-stage build — small final image, no Go toolchain shipped.
#
# Build context is the REPO ROOT (see docker-compose.yml) so the image can copy
# the shared libs (libs/go) and generated proto stubs (proto/gen) that the
# gateway's `replace` directives point at.
FROM golang:1.25-alpine AS build
WORKDIR /app

# Bring in the two local modules the gateway replaces, then its own source.
COPY libs/go/          ./libs/go/
COPY proto/gen/        ./proto/gen/
COPY services/gateway/ ./services/gateway/

WORKDIR /app/services/gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -o /server .

FROM alpine:3.20
RUN apk add --no-cache curl
COPY --from=build /server /server
EXPOSE 8080
CMD ["/server"]
