package main

import (
	"net/http"

	authpb "github.com/mashmool0/inama/proto/gen/auth"
)

// Auth handlers are the only public ones (no JWT). They forward the JSON body
// straight to the Auth gRPC service and return the resulting token pair.

func authRegisterHandler(client authpb.AuthServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req authpb.RegisterRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		pair, err := client.Register(r.Context(), &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, pair)
	})
}

func authLoginHandler(client authpb.AuthServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req authpb.LoginRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		pair, err := client.Login(r.Context(), &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, pair)
	})
}

func authRefreshHandler(client authpb.AuthServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req authpb.RefreshTokenRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		pair, err := client.RefreshToken(r.Context(), &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, pair)
	})
}

func authUpdateUsernameHandler(client authpb.AuthServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req authpb.UpdateUsernameRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		response, err := client.UpdateUsername(r.Context(), &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	})
}
