package main

import (
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mashmool0/inama/libs/identity"
)

// jwtAuth verifies the RS256 access token locally with Auth's public key and,
// on success, attaches the verified user id to the OUTGOING gRPC metadata via
// identity.WithUserID. The gateway NEVER calls Auth to verify — that keeps the
// hot path fast and Auth off the critical path for reads.
func jwtAuth(pubKey *rsa.PublicKey, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok || token == "" {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims := &jwt.RegisteredClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			// Reject anything that isn't RSA — defends against the alg-confusion
			// attack where a token is forged with HS256 using the public key as
			// the HMAC secret.
			if _, isRSA := t.Method.(*jwt.SigningMethodRSA); !isRSA {
				return nil, jwt.ErrSignatureInvalid
			}
			return pubKey, nil
		})
		if err != nil || !parsed.Valid || claims.Subject == "" {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := identity.WithUserID(r.Context(), claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
