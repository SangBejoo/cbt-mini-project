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
	base.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterMataPelajaranServiceHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterMateriServiceHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterTingkatServiceHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterSoalServiceHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterSoalDragDropServiceHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterTestSessionServiceHandlerFromEndpoint(ctx, mux, port, opts)
	base.RegisterHistoryServiceHandlerFromEndpoint(ctx, mux, port, opts)
}