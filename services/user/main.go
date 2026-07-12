package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	usergrpc "github.com/mashmool0/inama/services/user/internal/handler/grpc"
	"github.com/mashmool0/inama/services/user/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	sharedconfig "github.com/mashmool0/inama/libs/config"
	sharedidentity "github.com/mashmool0/inama/libs/identity"
	sharedlogging "github.com/mashmool0/inama/libs/logging"
)

func main() {
	cfg := sharedconfig.LoadBase("user")
	logger := sharedlogging.New(cfg.ServiceName)

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(sharedidentity.UnaryServerInterceptor()),
	)
	usergrpc.Register(grpcServer, service.NoopProfileService{}, service.NoopFollowService{})
	reflection.Register(grpcServer)

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = fmt.Fprintln(w, "user_service_info{service=\"user\"} 1")
	})
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"user"}`))
	})

	httpServer := &http.Server{
		Addr:              net.JoinHostPort("", strconv.Itoa(cfg.MetricsPort)),
		Handler:           httpMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	grpcListener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(cfg.GRPCPort)))
	if err != nil {
		logger.Error("failed to bind gRPC listener", "error", err)
		os.Exit(1)
	}

	errCh := make(chan error, 2)

	go func() {
		logger.Info("starting gRPC server", "port", cfg.GRPCPort)
		if serveErr := grpcServer.Serve(grpcListener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			errCh <- serveErr
		}
	}()

	go func() {
		logger.Info("starting HTTP server", "port", cfg.MetricsPort)
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")
	case serveErr := <-errCh:
		logger.Error("server stopped unexpectedly", "error", serveErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("failed to shut down HTTP server", "error", err)
	}

	logger.Info("user service stopped")
}
