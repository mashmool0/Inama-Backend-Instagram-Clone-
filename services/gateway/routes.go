package main

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// registerRoutes wires every REST path to its handler, wrapped in the correct
// middleware chain. Go 1.22 ServeMux gives us method + path-parameter matching
// ("POST /posts/{id}/like"), so no third-party router is needed.
func registerRoutes(mux *http.ServeMux, c clients, rdb *redis.Client, limit int, pubKey *rsa.PublicKey) {
	// Health — no middleware, used by Docker/monitoring probes.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "gateway"})
	})

	// public = rate limit only. Auth endpoints must be reachable without a token.
	public := func(h http.Handler) http.Handler {
		return rateLimit(rdb, limit, h)
	}
	// protected = rate limit + JWT verification + identity injection.
	protected := func(h http.Handler) http.Handler {
		return rateLimit(rdb, limit, jwtAuth(pubKey, h))
	}

	// ---- Auth (public) ----
	mux.Handle("POST /auth/register", public(authRegisterHandler(c.auth)))
	mux.Handle("POST /auth/login", public(authLoginHandler(c.auth)))
	mux.Handle("POST /auth/refresh", public(authRefreshHandler(c.auth)))
	mux.Handle("PATCH /auth/me/username", protected(authUpdateUsernameHandler(c.auth)))

	// ---- User (protected) ----
	mux.Handle("GET /users/{id}", protected(getUserHandler(c.user)))
	mux.Handle("GET /users/{id}/{action}", protected(getUserSubresourceHandler(c.user)))
	mux.Handle("PATCH /users/me", protected(updateUserHandler(c.user)))
	mux.Handle("POST /users/{id}/follow", protected(followHandler(c.user)))
	mux.Handle("DELETE /users/{id}/follow", protected(unfollowHandler(c.user)))

	// ---- Posts (protected) ----
	mux.Handle("POST /posts", protected(createPostHandler(c.posts)))
	mux.Handle("GET /posts/{id}", protected(getPostHandler(c.posts)))
	mux.Handle("DELETE /posts/{id}", protected(deletePostHandler(c.posts)))
	mux.Handle("POST /posts/{id}/like", protected(likePostHandler(c.posts)))
	mux.Handle("DELETE /posts/{id}/like", protected(unlikePostHandler(c.posts)))
	mux.Handle("POST /posts/{id}/comments", protected(addCommentHandler(c.posts)))
	mux.Handle("GET /posts/{id}/comments", protected(getCommentsHandler(c.posts)))

	// ---- Feed (protected) ----
	mux.Handle("GET /feed", protected(getFeedHandler(c.feed)))
	mux.Handle("GET /explore", protected(getExploreHandler(c.feed)))

	// ---- Notifications (protected) ----
	// Literal "read-all" beats the "{id}" wildcard in Go 1.22 ServeMux
	// regardless of registration order, so both can coexist.
	mux.Handle("GET /notifications", protected(getNotificationsHandler(c.notif)))
	mux.Handle("POST /notifications/read-all", protected(markAllReadHandler(c.notif)))
	mux.Handle("POST /notifications/{id}/read", protected(markReadHandler(c.notif)))

	// ---- Search (protected) ----
	mux.Handle("GET /search", protected(searchHandler(c.search)))
}
