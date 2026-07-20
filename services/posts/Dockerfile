# Multi-stage build — small final image, no Go toolchain shipped.
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server .

FROM alpine:3.20
RUN apk add --no-cache curl
COPY --from=build /server /server
EXPOSE 8080
CMD ["/server"]
