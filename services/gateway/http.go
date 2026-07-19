package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// maxBodyBytes caps request bodies so a client can't exhaust memory. 1 MiB is
// generous for the JSON payloads the gateway forwards.
const maxBodyBytes = 1 << 20

// decodeJSON reads and unmarshals the request body into dst, enforcing the size
// cap. Returns false (and writes a 400) if the body is missing or malformed, so
// callers can just `if !decodeJSON(...) { return }`.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
		}
		return false
	}
	return true
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError emits the gateway's uniform error envelope.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// queryInt parses an int query parameter, falling back to def when absent or
// unparseable. Used for pagination `limit`.
func queryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
