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

	sharedidentity "github.com/mashmool0/inama/libs/identity"
	sharedlogging "github.com/mashmool0/inama/libs/logging"
	notifconfig "github.com/mashmool0/inama/services/notifications/internal/config"
	notifgrpc "github.com/mashmool0/inama/services/notifications/internal/handler/grpc"
	"github.com/mashmool0/inama/services/notifications/internal/push"
	"github.com/mashmool0/inama/services/notifications/internal/queue"
	"github.com/mashmool0/inama/services/notifications/internal/repository"
	"github.com/mashmool0/inama/services/notifications/internal/schema"
	"github.com/mashmool0/inama/services/notifications/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := notifconfig.Load()
	logger := sharedlogging.New(cfg.ServiceName)

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repository.OpenPool(rootCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.AutoBootstrapSchema {
		if err := schema.Bootstrap(rootCtx, pool); err != nil {
			logger.Error("failed to bootstrap schema", "error", err)
			os.Exit(1)
		}
	}

	broker, err := queue.Connect(rootCtx, queue.Config{
		URL:         cfg.RabbitMQURL,
		Exchange:    cfg.RabbitMQExchange,
		QueueName:   cfg.RabbitMQQueue,
		RoutingKeys: cfg.RabbitMQRoutingKeys,
	})
	if err != nil {
		logger.Error("failed to connect to rabbitmq", "error", err)
		os.Exit(1)
	}
	defer broker.Close()

	notificationRepo := repository.NewNotificationRepository(pool)
	processedEventRepo := repository.NewProcessedEventRepository(pool)
	pushClient := push.NopClient{}
	readService := service.NewReadService(notificationRepo, cfg.DefaultPageLimit, cfg.MaxPageLimit)
	processor := service.NewProcessor(pool, notificationRepo, processedEventRepo, pushClient)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(sharedidentity.UnaryServerInterceptor()),
	)
	notifgrpc.Register(grpcServer, readService)
	reflection.Register(grpcServer)

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = fmt.Fprintln(w, "notifications_service_info{service=\"notifications\"} 1")
	})
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"notifications"}`))
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

	worker := queue.NewWorker(broker, processor)
	errCh := make(chan error, 3)

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

	go func() {
		logger.Info("starting queue worker shell")
		if err := worker.Run(rootCtx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")
	case serveErr := <-errCh:
		logger.Error("notifications service stopped unexpectedly", "error", serveErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("failed to shut down HTTP server", "error", err)
	}

	logger.Info("notifications service stopped")
}
