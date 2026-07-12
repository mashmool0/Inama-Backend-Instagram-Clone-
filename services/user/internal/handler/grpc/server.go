package grpc

import (
	"context"

	userpb "github.com/mashmool0/inama/proto/gen/user"
	"github.com/mashmool0/inama/services/user/internal/service"
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

func (s *Server) GetProfile(context.Context, *userpb.GetProfileRequest) (*userpb.Profile, error) {
	return nil, status.Error(codes.Unimplemented, "method GetProfile not implemented")
}

func (s *Server) UpdateProfile(context.Context, *userpb.UpdateProfileRequest) (*userpb.Profile, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateProfile not implemented")
}

func (s *Server) Follow(context.Context, *userpb.FollowRequest) (*userpb.FollowResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Follow not implemented")
}

func (s *Server) Unfollow(context.Context, *userpb.UnfollowRequest) (*userpb.UnfollowResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Unfollow not implemented")
}

func (s *Server) GetFollowers(context.Context, *userpb.GetFollowersRequest) (*userpb.UserIdPage, error) {
	return nil, status.Error(codes.Unimplemented, "method GetFollowers not implemented")
}

func (s *Server) GetFollowing(context.Context, *userpb.GetFollowingRequest) (*userpb.UserIdPage, error) {
	return nil, status.Error(codes.Unimplemented, "method GetFollowing not implemented")
}
