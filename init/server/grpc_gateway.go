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
	gwMux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Register your services here
	dependency.InitRestGatewayDependency(gwMux, opts, ctx, cfg)

	// Create a custom mux to handle both API and static files
	mux := http.NewServeMux()

	// Serve static files from uploads directory
	uploadsDir := "./uploads"
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// Serve API through gRPC-Gateway
	mux.Handle("/", gwMux)

	// Wrap mux with CORS middleware
	fmt.Printf("Starting HTTP server on port %d\n", cfg.RestServer.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.RestServer.Port), corsMiddleware(mux))
}