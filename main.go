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
	infraRedis "cbt-test-mini-project/init/infra/redis"
	"cbt-test-mini-project/init/logger"
	"cbt-test-mini-project/init/server"
	"cbt-test-mini-project/internal/event"
	authRepo "cbt-test-mini-project/internal/repository/auth"
	classRepo "cbt-test-mini-project/internal/repository/class"
	classStudentRepo "cbt-test-mini-project/internal/repository/class_student"
	mataPelajaranRepo "cbt-test-mini-project/internal/repository/mata_pelajaran"
	materiRepo "cbt-test-mini-project/internal/repository/materi"
	testSessionRepo "cbt-test-mini-project/internal/repository/test_session"
	tingkatRepo "cbt-test-mini-project/internal/repository/tingkat"
	"cbt-test-mini-project/internal/sync"
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

	// Initialize Redis for sync worker
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // Default DB

	if err := infraRedis.InitRedis(redisAddr, redisPassword, redisDB); err != nil {
		slog.Error("failed to initialize redis", "error", err)
		// Continue without sync - not fatal for CBT to run standalone
	} else {
		slog.Info("✓ Redis connected successfully")
	}
	defer func() {
		if err := infraRedis.CloseRedis(); err != nil {
			slog.Error("failed to close redis", "error", err)
		}
	}()

	// Initialize APM monitoring
	infra.InitAPM(*cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Sync Worker to listen for LMS events
	if infraRedis.RedisClient != nil {
		// Initialize repositories for sync worker
		materiRepoImpl := materiRepo.NewMateriRepository(repo.SQLDB)
		tingkatRepoImpl := tingkatRepo.NewTingkatRepository(repo.SQLDB)
		subjectRepoImpl := mataPelajaranRepo.NewMataPelajaranRepository(repo.SQLDB)
		authRepoImpl := authRepo.NewAuthRepository(repo.SQLDB)
		testSessionRepoImpl := testSessionRepo.NewTestSessionRepository(repo.SQLDB)
		classRepoImpl := classRepo.NewClassRepository(repo.SQLDB)
		classStudentRepoImpl := classStudentRepo.NewClassStudentRepository(repo.SQLDB)

		syncWorker := sync.NewSyncWorker(
			materiRepoImpl,
			tingkatRepoImpl,
			subjectRepoImpl,
			authRepoImpl,
			testSessionRepoImpl,
			classRepoImpl,
			classStudentRepoImpl,
		)
		go syncWorker.Start(ctx)
	}

	// Initialize Event Publisher
	publisher := event.NewPublisher(infraRedis.RedisClient)

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

