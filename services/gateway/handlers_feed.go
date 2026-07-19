package main

import (
	"net/http"

	feedpb "github.com/mashmool0/inama/proto/gen/feed"
)

// Feed is per-user: the Feed service derives whose feed to build from x-user-id.

func getFeedHandler(client feedpb.FeedServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := client.GetFeed(r.Context(), &feedpb.GetFeedRequest{
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

func getExploreHandler(client feedpb.FeedServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := client.GetExplore(r.Context(), &feedpb.GetExploreRequest{
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
