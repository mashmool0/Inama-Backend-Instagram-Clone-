package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authpb "github.com/mashmool0/inama/proto/gen/auth"
	notifpb "github.com/mashmool0/inama/proto/gen/notif"
	userpb "github.com/mashmool0/inama/proto/gen/user"
)

type fakeAuthServer struct {
	authpb.UnimplementedAuthServiceServer
	callerID string
}

func (s *fakeAuthServer) Register(_ context.Context, req *authpb.RegisterRequest) (*authpb.TokenPair, error) {
	if req.GetUsername() == "taken" {
		return nil, status.Error(codes.AlreadyExists, "username already taken")
	}
	return &authpb.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}, nil
}

func (s *fakeAuthServer) Login(context.Context, *authpb.LoginRequest) (*authpb.TokenPair, error) {
	return &authpb.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}, nil
}

func (s *fakeAuthServer) RefreshToken(context.Context, *authpb.RefreshTokenRequest) (*authpb.TokenPair, error) {
	return &authpb.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh", ExpiresIn: 900}, nil
}

func (s *fakeAuthServer) UpdateUsername(ctx context.Context, req *authpb.UpdateUsernameRequest) (*authpb.UpdateUsernameResponse, error) {
	s.callerID = incomingUserID(ctx)
	return &authpb.UpdateUsernameResponse{Username: req.GetUsername()}, nil
}

type fakeUserServer struct {
	userpb.UnimplementedUserServiceServer
	callerID     string
	updated      *userpb.UpdateProfileRequest
	followTarget string
	pageLimit    int32
	pageCursor   string
}

func (s *fakeUserServer) GetProfile(_ context.Context, req *userpb.GetProfileRequest) (*userpb.Profile, error) {
	return &userpb.Profile{Id: req.GetUserId(), Username: "by-id"}, nil
}

func (s *fakeUserServer) GetProfileByUsername(_ context.Context, req *userpb.GetProfileByUsernameRequest) (*userpb.Profile, error) {
	if req.GetUsername() == "unavailable" {
		return nil, status.Error(codes.Unavailable, "offline")
	}
	return &userpb.Profile{Id: "profile-id", Username: req.GetUsername()}, nil
}

func (s *fakeUserServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.Profile, error) {
	s.callerID = incomingUserID(ctx)
	s.updated = req
	return &userpb.Profile{Id: s.callerID, Username: "alice", Bio: req.GetBio(), AvatarUrl: req.GetAvatarUrl()}, nil
}

func (s *fakeUserServer) Follow(ctx context.Context, req *userpb.FollowRequest) (*userpb.FollowResponse, error) {
	s.callerID = incomingUserID(ctx)
	s.followTarget = req.GetTargetUserId()
	return &userpb.FollowResponse{Success: true}, nil
}

func (s *fakeUserServer) Unfollow(ctx context.Context, req *userpb.UnfollowRequest) (*userpb.UnfollowResponse, error) {
	s.callerID = incomingUserID(ctx)
	s.followTarget = req.GetTargetUserId()
	return &userpb.UnfollowResponse{Success: true}, nil
}

func (s *fakeUserServer) GetFollowers(_ context.Context, req *userpb.GetFollowersRequest) (*userpb.UserIdPage, error) {
	s.pageLimit, s.pageCursor = req.GetLimit(), req.GetCursor()
	return &userpb.UserIdPage{UserIds: []string{"follower"}, NextCursor: "next"}, nil
}

func (s *fakeUserServer) GetFollowing(_ context.Context, req *userpb.GetFollowingRequest) (*userpb.UserIdPage, error) {
	s.pageLimit, s.pageCursor = req.GetLimit(), req.GetCursor()
	return &userpb.UserIdPage{UserIds: []string{"following"}}, nil
}

type fakeNotificationsServer struct {
	notifpb.UnimplementedNotificationsServiceServer
	callerID string
	markedID string
	limit    int32
	cursor   string
}

func (s *fakeNotificationsServer) GetNotifications(ctx context.Context, req *notifpb.GetNotificationsRequest) (*notifpb.NotificationPage, error) {
	s.callerID, s.limit, s.cursor = incomingUserID(ctx), req.GetLimit(), req.GetCursor()
	return &notifpb.NotificationPage{Notifications: []*notifpb.Notification{{Id: "notification-id"}}}, nil
}

func (s *fakeNotificationsServer) MarkAsRead(ctx context.Context, req *notifpb.MarkAsReadRequest) (*notifpb.MarkResponse, error) {
	s.callerID, s.markedID = incomingUserID(ctx), req.GetNotificationId()
	if req.GetNotificationId() == "foreign" {
		return nil, status.Error(codes.NotFound, "notification not found")
	}
	return &notifpb.MarkResponse{Success: true}, nil
}

func (s *fakeNotificationsServer) MarkAllRead(ctx context.Context, _ *notifpb.MarkAllReadRequest) (*notifpb.MarkResponse, error) {
	s.callerID = incomingUserID(ctx)
	return &notifpb.MarkResponse{Success: true}, nil
}

type gatewayFixture struct {
	handler http.Handler
	key     *rsa.PrivateKey
	auth    *fakeAuthServer
	user    *fakeUserServer
	notif   *fakeNotificationsServer
}

func newGatewayFixture(t *testing.T) gatewayFixture {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	authServer := &fakeAuthServer{}
	userServer := &fakeUserServer{}
	notifServer := &fakeNotificationsServer{}
	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, authServer)
	userpb.RegisterUserServiceServer(grpcServer, userServer)
	notifpb.RegisterNotificationsServiceServer(grpcServer, notifServer)
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	mux := http.NewServeMux()
	registerRoutes(mux, clients{
		auth:  authpb.NewAuthServiceClient(conn),
		user:  userpb.NewUserServiceClient(conn),
		notif: notifpb.NewNotificationsServiceClient(conn),
	}, nil, 0, &key.PublicKey)

	return gatewayFixture{
		handler: cors([]string{"http://localhost:3000"}, mux),
		key:     key,
		auth:    authServer,
		user:    userServer,
		notif:   notifServer,
	}
}

func TestPublicAuthRoutesAndMalformedJSON(t *testing.T) {
	fixture := newGatewayFixture(t)
	tests := []struct {
		path   string
		body   string
		status int
	}{
		{"/auth/register", `{"email":"a@x.com","username":"alice","password":"pw"}`, http.StatusCreated},
		{"/auth/login", `{"identifier":"alice","password":"pw"}`, http.StatusOK},
		{"/auth/refresh", `{"refresh_token":"token"}`, http.StatusOK},
		{"/auth/register", `{broken`, http.StatusBadRequest},
		{"/auth/register", `{"email":"b@x.com","username":"taken","password":"pw"}`, http.StatusConflict},
	}

	for _, test := range tests {
		recorder := performRequest(fixture.handler, http.MethodPost, test.path, test.body, "")
		if recorder.Code != test.status {
			t.Errorf("POST %s status = %d, want %d; body=%s", test.path, recorder.Code, test.status, recorder.Body.String())
		}
	}
}

func TestProtectedRoutesRequireJWTAndForwardSubject(t *testing.T) {
	fixture := newGatewayFixture(t)
	if got := performRequest(fixture.handler, http.MethodGet, "/users/profile-id", "", "").Code; got != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", got, http.StatusUnauthorized)
	}

	token := accessToken(t, fixture.key, "actor-id")
	response := performRequest(fixture.handler, http.MethodPatch, "/auth/me/username", `{"username":"new_name"}`, token)
	if response.Code != http.StatusOK || fixture.auth.callerID != "actor-id" {
		t.Fatalf("username update status=%d caller=%q body=%s", response.Code, fixture.auth.callerID, response.Body.String())
	}

	response = performRequest(fixture.handler, http.MethodPatch, "/users/me", `{"bio":"hello","avatar_url":"avatar"}`, token)
	if response.Code != http.StatusOK || fixture.user.callerID != "actor-id" {
		t.Fatalf("profile update status=%d caller=%q body=%s", response.Code, fixture.user.callerID, response.Body.String())
	}
	if fixture.user.updated.Username != nil || fixture.user.updated.GetBio() != "hello" || fixture.user.updated.GetAvatarUrl() != "avatar" {
		t.Fatalf("forwarded profile update = %+v", fixture.user.updated)
	}

	response = performRequest(fixture.handler, http.MethodPatch, "/users/me", `{"username":"forbidden"}`, token)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("direct username update status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestUserAndNotificationRoutesForwardParameters(t *testing.T) {
	fixture := newGatewayFixture(t)
	token := accessToken(t, fixture.key, "actor-id")

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/users/profile-id"},
		{http.MethodGet, "/users/username/alice"},
		{http.MethodPost, "/users/target-id/follow"},
		{http.MethodDelete, "/users/target-id/follow"},
		{http.MethodGet, "/users/target-id/followers?limit=7&cursor=edge"},
		{http.MethodGet, "/users/target-id/following?limit=8&cursor=next-edge"},
		{http.MethodGet, "/notifications?limit=9&cursor=notif-edge"},
		{http.MethodPost, "/notifications/notification-id/read"},
		{http.MethodPost, "/notifications/read-all"},
	}
	for _, test := range tests {
		response := performRequest(fixture.handler, test.method, test.path, "", token)
		if response.Code != http.StatusOK {
			t.Fatalf("%s %s status=%d body=%s", test.method, test.path, response.Code, response.Body.String())
		}
	}

	if fixture.user.followTarget != "target-id" || fixture.user.pageLimit != 8 || fixture.user.pageCursor != "next-edge" {
		t.Fatalf("user parameters target=%q limit=%d cursor=%q", fixture.user.followTarget, fixture.user.pageLimit, fixture.user.pageCursor)
	}
	if fixture.notif.callerID != "actor-id" || fixture.notif.markedID != "notification-id" || fixture.notif.limit != 9 || fixture.notif.cursor != "notif-edge" {
		t.Fatalf("notification parameters caller=%q id=%q limit=%d cursor=%q", fixture.notif.callerID, fixture.notif.markedID, fixture.notif.limit, fixture.notif.cursor)
	}
}

func TestGatewayErrorMappingAndCORS(t *testing.T) {
	fixture := newGatewayFixture(t)
	token := accessToken(t, fixture.key, "actor-id")

	if got := performRequest(fixture.handler, http.MethodGet, "/users/username/unavailable", "", token).Code; got != http.StatusServiceUnavailable {
		t.Fatalf("unavailable status = %d, want %d", got, http.StatusServiceUnavailable)
	}
	if got := performRequest(fixture.handler, http.MethodPost, "/notifications/foreign/read", "", token).Code; got != http.StatusNotFound {
		t.Fatalf("foreign notification status = %d, want %d", got, http.StatusNotFound)
	}

	request := httptest.NewRequest(http.MethodOptions, "/notifications", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	recorder := httptest.NewRecorder()
	fixture.handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent || recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("preflight status=%d origin=%q", recorder.Code, recorder.Header().Get("Access-Control-Allow-Origin"))
	}

	request = httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	recorder = httptest.NewRecorder()
	fixture.handler.ServeHTTP(recorder, request)
	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("normal response is missing CORS origin")
	}
}

func performRequest(handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func accessToken(t *testing.T, key *rsa.PrivateKey, subject string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(key)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return token
}

func incomingUserID(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get("x-user-id")
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func decodeResponse[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	return value
}
