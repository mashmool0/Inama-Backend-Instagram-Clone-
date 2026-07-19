package main

import (
	"net/http"

	notifpb "github.com/mashmool0/inama/proto/gen/notif"
)

// Notifications belong to the authenticated user (x-user-id); the recipient is
// never taken from the request.

func getNotificationsHandler(client notifpb.NotificationsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := client.GetNotifications(r.Context(), &notifpb.GetNotificationsRequest{
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

func markReadHandler(client notifpb.NotificationsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.MarkAsRead(r.Context(), &notifpb.MarkAsReadRequest{
			NotificationId: r.PathValue("id"),
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
}

func markAllReadHandler(client notifpb.NotificationsServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.MarkAllRead(r.Context(), &notifpb.MarkAllReadRequest{})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	})
}
