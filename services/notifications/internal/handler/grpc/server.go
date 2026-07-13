package grpc

import (
	"context"

	notifpb "github.com/mashmool0/inama/proto/gen/notif"
	"github.com/mashmool0/inama/services/notifications/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	notifpb.UnimplementedNotificationsServiceServer

	reads service.ReadService
}

func Register(registrar grpc.ServiceRegistrar, reads service.ReadService) {
	notifpb.RegisterNotificationsServiceServer(registrar, &Server{reads: reads})
}

func (s *Server) GetNotifications(context.Context, *notifpb.GetNotificationsRequest) (*notifpb.NotificationPage, error) {
	return nil, status.Error(codes.Unimplemented, "method GetNotifications not implemented")
}

func (s *Server) MarkAsRead(context.Context, *notifpb.MarkAsReadRequest) (*notifpb.MarkResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method MarkAsRead not implemented")
}

func (s *Server) MarkAllRead(context.Context, *notifpb.MarkAllReadRequest) (*notifpb.MarkResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method MarkAllRead not implemented")
}
