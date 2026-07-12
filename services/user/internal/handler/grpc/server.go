package grpc

import (
	"context"
	"errors"

	"github.com/mashmool0/inama/libs/identity"
	userpb "github.com/mashmool0/inama/proto/gen/user"
	"github.com/mashmool0/inama/services/user/internal/model"
	"github.com/mashmool0/inama/services/user/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	userpb.UnimplementedUserServiceServer

	profiles service.ProfileService
	follows  service.FollowService
}

func Register(registrar grpc.ServiceRegistrar, profiles service.ProfileService, follows service.FollowService) {
	userpb.RegisterUserServiceServer(registrar, &Server{
		profiles: profiles,
		follows:  follows,
	})
}

func (s *Server) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.Profile, error) {
	profile, err := s.profiles.GetProfile(ctx, req.GetUserId())
	if err != nil {
		return nil, translateError(err)
	}

	return toProtoProfile(profile), nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.Profile, error) {
	actorID, err := identity.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.profiles.UpdateProfile(ctx, actorID, model.UpdateProfileInput{
		Username:  req.Username,
		Bio:       req.Bio,
		AvatarURL: req.AvatarUrl,
	})
	if err != nil {
		return nil, translateError(err)
	}

	return toProtoProfile(profile), nil
}

func (s *Server) Follow(ctx context.Context, req *userpb.FollowRequest) (*userpb.FollowResponse, error) {
	actorID, err := identity.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.follows.Follow(ctx, actorID, req.GetTargetUserId())
	if err != nil {
		return nil, translateError(err)
	}

	return &userpb.FollowResponse{Success: true}, nil
}

func (s *Server) Unfollow(ctx context.Context, req *userpb.UnfollowRequest) (*userpb.UnfollowResponse, error) {
	actorID, err := identity.RequireUserID(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.follows.Unfollow(ctx, actorID, req.GetTargetUserId())
	if err != nil {
		return nil, translateError(err)
	}

	return &userpb.UnfollowResponse{Success: true}, nil
}

func (s *Server) GetFollowers(ctx context.Context, req *userpb.GetFollowersRequest) (*userpb.UserIdPage, error) {
	page, err := s.follows.GetFollowers(ctx, req.GetUserId(), req.GetLimit(), req.GetCursor())
	if err != nil {
		return nil, translateError(err)
	}

	return &userpb.UserIdPage{
		UserIds:    page.UserIDs,
		NextCursor: page.NextCursor,
	}, nil
}

func (s *Server) GetFollowing(ctx context.Context, req *userpb.GetFollowingRequest) (*userpb.UserIdPage, error) {
	page, err := s.follows.GetFollowing(ctx, req.GetUserId(), req.GetLimit(), req.GetCursor())
	if err != nil {
		return nil, translateError(err)
	}

	return &userpb.UserIdPage{
		UserIds:    page.UserIDs,
		NextCursor: page.NextCursor,
	}, nil
}

func toProtoProfile(profile model.Profile) *userpb.Profile {
	return &userpb.Profile{
		Id:             profile.ID,
		Username:       profile.Username,
		Bio:            profile.Bio,
		AvatarUrl:      profile.AvatarURL,
		FollowerCount:  profile.FollowerCount,
		FollowingCount: profile.FollowingCount,
		CreatedAt:      timestamppb.New(profile.CreatedAt),
	}
}

func translateError(err error) error {
	switch {
	case errors.Is(err, service.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrInvalidCursor):
		return status.Error(codes.InvalidArgument, err.Error())
	}

	var invalidArgument service.InvalidArgumentError
	if errors.As(err, &invalidArgument) {
		return status.Error(codes.InvalidArgument, invalidArgument.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
