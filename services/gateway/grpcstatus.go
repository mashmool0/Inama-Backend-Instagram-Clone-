package main

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// writeGRPCError translates a downstream gRPC error into the matching HTTP
// status + JSON envelope. Internal/unknown failures are flattened to a generic
// 500 so we never leak backend implementation details to clients.
func writeGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	switch st.Code() {
	case codes.OK:
		return
	case codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		writeError(w, http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		writeError(w, http.StatusForbidden, st.Message())
	case codes.NotFound:
		writeError(w, http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		writeError(w, http.StatusConflict, st.Message())
	case codes.ResourceExhausted:
		writeError(w, http.StatusTooManyRequests, st.Message())
	case codes.Unimplemented:
		writeError(w, http.StatusNotImplemented, st.Message())
	case codes.Unavailable:
		writeError(w, http.StatusServiceUnavailable, "service unavailable")
	case codes.DeadlineExceeded:
		writeError(w, http.StatusGatewayTimeout, "upstream timeout")
	default: // Internal, Unknown, DataLoss, ...
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
