// The API Gateway is the single HTTP entrypoint for all clients. It:
//   - rate-limits per client IP (Redis fixed-window),
//   - verifies the RS256 JWT locally with Auth's PUBLIC key (never calls Auth),
//   - injects the verified user id into gRPC metadata (x-user-id),
//   - routes each REST path to the owning service over gRPC.
//
// Everything is package main: the gateway is a binary, not a library.
package main

import (
	"crypto/rsa"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mashmool0/inama/libs/config"
	"github.com/mashmool0/inama/libs/logging"

	authpb "github.com/mashmool0/inama/proto/gen/auth"
	feedpb "github.com/mashmool0/inama/proto/gen/feed"
	notifpb "github.com/mashmool0/inama/proto/gen/notif"
	postspb "github.com/mashmool0/inama/proto/gen/posts"
	searchpb "github.com/mashmool0/inama/proto/gen/search"
	userpb "github.com/mashmool0/inama/proto/gen/user"
)

// clients bundles one gRPC client per downstream service. Connections are lazy
// (grpc.NewClient) — a service being down surfaces as a 503 at call time, not
// at startup, so the gateway can boot before every backend is ready.
type clients struct {
	auth   authpb.AuthServiceClient
	user   userpb.UserServiceClient
	posts  postspb.PostsServiceClient
	feed   feedpb.FeedServiceClient
	notif  notifpb.NotificationsServiceClient
	search searchpb.SearchServiceClient
}

func main() {
	log := logging.New("gateway")

	port := config.Get("PORT", "8080")
	pubKeyPath := config.Get("JWT_PUBLIC_KEY_PATH", "/keys/jwt_public.pem")
	redisAddr := config.Get("REDIS_ADDR", "redis:6379")
	rateLimit := config.GetInt("RATE_LIMIT_RPS", 100)

	pubKey := loadPublicKey(log, pubKeyPath)

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	dial := func(addr string) *grpc.ClientConn {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Error("failed to create gRPC client", "addr", addr, "err", err)
			os.Exit(1)
		}
		return conn
	}

	c := clients{
		auth:   authpb.NewAuthServiceClient(dial(config.Get("AUTH_ADDR", "auth:50051"))),
		user:   userpb.NewUserServiceClient(dial(config.Get("USER_ADDR", "user:50051"))),
		posts:  postspb.NewPostsServiceClient(dial(config.Get("POSTS_ADDR", "posts:50051"))),
		feed:   feedpb.NewFeedServiceClient(dial(config.Get("FEED_ADDR", "feed:50051"))),
		notif:  notifpb.NewNotificationsServiceClient(dial(config.Get("NOTIF_ADDR", "notifications:50051"))),
		search: searchpb.NewSearchServiceClient(dial(config.Get("SEARCH_ADDR", "search:50051"))),
	}

	mux := http.NewServeMux()
	registerRoutes(mux, c, rdb, rateLimit, pubKey)
	handler := cors(strings.Split(config.Get("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","), mux)

	log.Info("gateway listening", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Error("server exited", "err", err)
		os.Exit(1)
	}
}

// loadPublicKey reads Auth's RS256 public key from disk. On a cold start the
// gateway boots before Auth has waited for Postgres, generated the shared key,
// and run migrations — which takes well over 20s on an empty volume — so it
// retries for up to two minutes before giving up. Paired with a restart policy
// in compose, this makes the boot race non-fatal.
func loadPublicKey(log logger, path string) *rsa.PublicKey {
	for attempt := 1; attempt <= 60; attempt++ {
		pem, err := os.ReadFile(path)
		if err == nil {
			key, perr := jwt.ParseRSAPublicKeyFromPEM(pem)
			if perr == nil {
				log.Info("loaded JWT public key", "path", path)
				return key
			}
			log.Error("public key file is not valid PEM", "path", path, "err", perr)
			os.Exit(1)
		}
		log.Warn("JWT public key not ready, retrying", "path", path, "attempt", attempt)
		time.Sleep(2 * time.Second)
	}
	log.Error("JWT public key never appeared", "path", path)
	os.Exit(1)
	return nil
}

// logger is the subset of *slog.Logger the startup path needs.
type logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}
