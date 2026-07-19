package main

import (
	"net/http"

	searchpb "github.com/mashmool0/inama/proto/gen/search"
)

// searchTypes maps the ?type= query value to its proto enum. Absent/unknown =>
// UNSPECIFIED, which the Search service treats as "all".
var searchTypes = map[string]searchpb.SearchType{
	"user":    searchpb.SearchType_SEARCH_TYPE_USER,
	"hashtag": searchpb.SearchType_SEARCH_TYPE_HASHTAG,
	"text":    searchpb.SearchType_SEARCH_TYPE_TEXT,
}

func searchHandler(client searchpb.SearchServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		resp, err := client.Search(r.Context(), &searchpb.SearchRequest{
			Query:  q.Get("q"),
			Type:   searchTypes[q.Get("type")], // zero value = UNSPECIFIED
			Limit:  int32(queryInt(r, "limit", 20)),
			Cursor: q.Get("cursor"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
}
