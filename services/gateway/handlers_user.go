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

func getUserSubresourceHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, action := r.PathValue("id"), r.PathValue("action")
		switch {
		case id == "username":
			profile, err := client.GetProfileByUsername(r.Context(), &userpb.GetProfileByUsernameRequest{Username: action})
			if err != nil {
				writeGRPCError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, profile)
		case action == "followers":
			getFollowers(w, r, client, id)
		case action == "following":
			getFollowing(w, r, client, id)
		default:
			writeError(w, http.StatusNotFound, "route not found")
		}
	})
}

func updateUserHandler(client userpb.UserServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Bio       *string `json:"bio"`
			AvatarURL *string `json:"avatar_url"`
		}
		if !decodeJSON(w, r, &body) {
			return
		}
		req := userpb.UpdateProfileRequest{Bio: body.Bio, AvatarUrl: body.AvatarURL}
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

func getFollowers(w http.ResponseWriter, r *http.Request, client userpb.UserServiceClient, userID string) {
	page, err := client.GetFollowers(r.Context(), &userpb.GetFollowersRequest{
		UserId: userID,
		Limit:  int32(queryInt(r, "limit", 20)),
		Cursor: r.URL.Query().Get("cursor"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func getFollowing(w http.ResponseWriter, r *http.Request, client userpb.UserServiceClient, userID string) {
	page, err := client.GetFollowing(r.Context(), &userpb.GetFollowingRequest{
		UserId: userID,
		Limit:  int32(queryInt(r, "limit", 20)),
		Cursor: r.URL.Query().Get("cursor"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}
