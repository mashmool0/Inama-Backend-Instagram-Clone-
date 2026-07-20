// Package identity carries the authenticated user across gRPC calls via
// metadata (the "x-user-id" header). This is the ONE place the identity
// convention lives — every Go service uses it, so no service re-invents auth.
//
// Flow:
//
//	Gateway verifies the JWT, then WithUserID(ctx, id) on the OUTGOING call.
//	Service registers UnaryServerInterceptor(); handlers read UserIDFromContext(ctx).
//
// Security note: services trust this value because they are only reachable from
// the Gateway on the private network, and the Gateway overwrites any client-set
// x-user-id with the id it extracted from the verified JWT.
package identity

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MetadataKey is the metadata field the verified user id travels in.
const MetadataKey = "x-user-id"

type contextKey struct{}

var userIDKey = contextKey{}

// WithUserID attaches the user id to an OUTGOING context. The Gateway calls this
// after verifying the JWT, just before invoking a downstream service.
func WithUserID(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, MetadataKey, userID)
}

// UserIDFromContext returns the authenticated user id a handler is serving, and
// whether one was present. Use this instead of trusting any request field.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok && v != ""
}

// RequireUserID is a convenience for handlers that must have an authenticated
// caller. Returns an Unauthenticated gRPC error if the id is missing.
func RequireUserID(ctx context.Context) (string, error) {
	if id, ok := UserIDFromContext(ctx); ok {
		return id, nil
	}
	return "", status.Error(codes.Unauthenticated, "missing authenticated user")
}

// UnaryServerInterceptor reads x-user-id from incoming metadata and puts it on
// the context so handlers can read it with UserIDFromContext.
//
//	server := grpc.NewServer(grpc.UnaryInterceptor(identity.UnaryServerInterceptor()))
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get(MetadataKey); len(vals) > 0 && vals[0] != "" {
				ctx = context.WithValue(ctx, userIDKey, vals[0])
			}
		}
		return handler(ctx, req)
	}
}
