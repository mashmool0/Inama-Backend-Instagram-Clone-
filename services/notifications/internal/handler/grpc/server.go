package grpc

import (
	"context"
	"errors"

	"github.com/mashmool0/inama/libs/identity"
	notifpb "github.com/mashmool0/inama/proto/gen/notif"
	"github.com/mashmool0/inama/services/notifications/internal/model"
	"github.com/mashmool0/inama/services/notifications/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	notifpb.UnimplementedNotificationsServiceServer

	reads service.ReadService
}

func Register(registrar grpc.ServiceRegistrar, reads service.ReadService) {
	notifpb.RegisterNotificationsServiceServer(registrar, &Server{reads: reads})
}

func (s *Server) GetNotifications(ctx context.Context, req *notifpb.GetNotificationsRequest) (*notifpb.NotificationPage, error) {
	recipientID, err := identity.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.reads.GetNotifications(ctx, recipientID, req.GetLimit(), req.GetCursor())
	if err != nil {
		return nil, translateError(err)
	}

	protoNotifications := make([]*notifpb.Notification, 0, len(page.Notifications))
	for _, notification := range page.Notifications {
		protoNotifications = append(protoNotifications, toProtoNotification(notification))
	}

	return &notifpb.NotificationPage{
		Notifications: protoNotifications,
		NextCursor:    page.NextCursor,
	}, nil
}

func (s *Server) MarkAsRead(ctx context.Context, req *notifpb.MarkAsReadRequest) (*notifpb.MarkResponse, error) {
	recipientID, err := identity.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.reads.MarkAsRead(ctx, recipientID, req.GetNotificationId()); err != nil {
		return nil, translateError(err)
	}

	return &notifpb.MarkResponse{Success: true}, nil
}

func (s *Server) MarkAllRead(ctx context.Context, _ *notifpb.MarkAllReadRequest) (*notifpb.MarkResponse, error) {
	recipientID, err := identity.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.reads.MarkAllRead(ctx, recipientID); err != nil {
		return nil, translateError(err)
	}

	return &notifpb.MarkResponse{Success: true}, nil
}

func toProtoNotification(notification model.Notification) *notifpb.Notification {
	var postID string
	if notification.PostID != nil {
		postID = *notification.PostID
	}

	return &notifpb.Notification{
		Id:        notification.ID,
		Type:      notifpb.NotificationType(notification.Type),
		ActorId:   notification.ActorID,
		PostId:    postID,
		IsRead:    notification.IsRead,
		CreatedAt: timestamppb.New(notification.CreatedAt),
	}
}

func translateError(err error) error {
	switch {
	case errors.Is(err, service.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrInvalidCursor):
		return status.Error(codes.InvalidArgument, err.Error())
	}

	var invalidArgument service.InvalidArgumentError
	if errors.As(err, &invalidArgument) {
		return status.Error(codes.InvalidArgument, invalidArgument.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
