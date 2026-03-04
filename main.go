package main

import (
	"context"
	"fmt"
	"log/slog"
	_ "net/http/pprof"
	"os"
	"time"

	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra"
	"cbt-test-mini-project/init/logger"
	"cbt-test-mini-project/init/server"
	"cbt-test-mini-project/internal/event"
	"cbt-test-mini-project/util"
)

var cfg *config.Main

func init() {
	cfg = config.Load()
	logger.Load(*cfg)
}

func main() {
	// Load repository
	repo := infra.LoadRepository(*cfg)
	defer func() {
		if errClose := repo.Close(); errClose != nil {
			slog.Error("failed to close repositories", "error", errClose)
		}
	}()

	// Initialize APM monitoring
	infra.InitAPM(*cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Event Publisher
	publisher := event.NewPublisher(nil)
	outboxWorker := event.NewOutboxWorker(repo.SQLDB)
	go outboxWorker.Start(ctx)

	grpcServer, err := server.RunGRPCServer(ctx, *cfg, *repo, publisher)
	if err != nil {
		slog.Error("failed to run grpc server", "error", err)
		os.Exit(1)
	}

	restServer, err := server.RunGatewayRestServer(ctx, *cfg, *repo, publisher)

	if err != nil {
		slog.Error("failed to run gateway rest server", "error", err)
		os.Exit(1)
	}

	slog.Info("servers started successfully", "grpc_port", cfg.GrpcServer.Port, "rest_port", cfg.RestServer.Port)

	// Print startup banner
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("               🚀 CBT Mini Project Server Started 🚀")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("  gRPC Server   : http://localhost:%d\n", cfg.GrpcServer.Port)
	fmt.Printf("  REST Gateway  : http://localhost:%d\n", cfg.RestServer.Port)
	fmt.Printf("  APM Dashboard : http://localhost:5601\n")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	wait := util.GracefulShutdown(ctx, 30*time.Second, map[string]util.Operation{
		"grpc": func(ctx context.Context) error {
			grpcServer.GracefulStop()
			slog.Info("grpc server stopped gracefully")
			return nil
		},
		"rest_gateway": func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := restServer.Shutdown(shutdownCtx); err != nil {
				slog.Error("rest gateway shutdown failed", "error", err)
				return err
			}
			slog.Info("rest gateway stopped gracefully")
			return nil
		},
	})
	<-wait
	slog.Info("application shutdown complete")
}
