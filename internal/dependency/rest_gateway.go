package dependency

import (
	"context"
	"fmt"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/init/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func InitRestGatewayDependency(mux *runtime.ServeMux, opts []grpc.DialOption, ctx context.Context, cfg config.Main) {
	port := fmt.Sprintf(":%d", cfg.GrpcServer.Port)
	base.RegisterBaseHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterNotesServiceHandlerFromEndpoint(ctx, mux, port, opts)
}