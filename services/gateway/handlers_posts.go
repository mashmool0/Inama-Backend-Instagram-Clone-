package main

import (
	"net/http"

	postspb "github.com/mashmool0/inama/proto/gen/posts"
)

// The author of a post/comment/like is the authenticated user, resolved by the
// Posts service from x-user-id metadata — never trusted from the body.

func createPostHandler(client postspb.PostsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req postspb.CreatePostRequest // {caption, media_url}
		if !decodeJSON(w, r, &req) {
			return
		}
		post, err := client.CreatePost(r.Context(), &req)
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, post)
	})
}

func getPostHandler(client postspb.PostsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		post, err := client.GetPost(r.Context(), &postspb.GetPostRequest{
			PostId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, post)
	})
}

func deletePostHandler(client postspb.PostsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := client.DeletePost(r.Context(), &postspb.DeletePostRequest{
			PostId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func likePostHandler(client postspb.PostsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.LikePost(r.Context(), &postspb.LikePostRequest{
			PostId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

func unlikePostHandler(client postspb.PostsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.UnlikePost(r.Context(), &postspb.UnlikePostRequest{
			PostId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

func addCommentHandler(client postspb.PostsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Body string `json:"body"`
		}
		if !decodeJSON(w, r, &body) {
			return
		}
		comment, err := client.AddComment(r.Context(), &postspb.AddCommentRequest{
			PostId: r.PathValue("id"),
			Body:   body.Body,
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, comment)
	})
}

func getCommentsHandler(client postspb.PostsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := client.GetComments(r.Context(), &postspb.GetCommentsRequest{
			PostId: r.PathValue("id"),
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
