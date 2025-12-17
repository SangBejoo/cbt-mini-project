package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Update this import path
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra"
	"cbt-test-mini-project/internal/dependency"
)

func RunGatewayRestServer(ctx context.Context, cfg config.Main, repo infra.Repository) error {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Register your services here
	dependency.InitRestGatewayDependency(mux, opts, ctx, cfg)

	// Wrap mux with CORS middleware
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.RestServer.Port), corsMiddleware(mux))
}
