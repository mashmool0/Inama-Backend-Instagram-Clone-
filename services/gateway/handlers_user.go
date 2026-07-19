package main

import (
	"net/http"

	userpb "github.com/mashmool0/inama/proto/gen/user"
)

// The follower/current user is never taken from the request — the User service
// reads it from the x-user-id metadata the JWT middleware injected. Handlers
// only pass the *target* id (from the path).

func getUserHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile, err := client.GetProfile(r.Context(), &userpb.GetProfileRequest{
			UserId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, profile)
	})
}

func updateUserHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req userpb.UpdateProfileRequest // fields are optional (*string)
		if !decodeJSON(w, r, &req) {
			return
		}
		profile, err := client.UpdateProfile(r.Context(), &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, profile)
	})
}

func followHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Follow(r.Context(), &userpb.FollowRequest{
			TargetUserId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

func unfollowHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Unfollow(r.Context(), &userpb.UnfollowRequest{
			TargetUserId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

func getFollowersHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := client.GetFollowers(r.Context(), &userpb.GetFollowersRequest{
			UserId: r.PathValue("id"),
			Limit:  int32(queryInt(r, "limit", 20)),
			Cursor: r.URL.Query().Get("cursor"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
	})
}

func getFollowingHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := client.GetFollowing(r.Context(), &userpb.GetFollowingRequest{
			UserId: r.PathValue("id"),
			Limit:  int32(queryInt(r, "limit", 20)),
			Cursor: r.URL.Query().Get("cursor"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
	})
}
