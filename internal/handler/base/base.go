package handler

import (
	"context"

	base "cbt-test-mini-project/gen/proto"

	"go.elastic.co/apm/v2"
	"google.golang.org/protobuf/types/known/emptypb"
)
func (h *baseHandler) HealthCheck(ctx context.Context, request *emptypb.Empty) (response *base.MessageStatusResponse, err error) {

	span, ctx := apm.StartSpan(ctx, "transport.HealthCheck", "transport.internal")
	if span != nil {
		defer span.End()
	}

	response = &base.MessageStatusResponse{
		Status:  "OK",
		Message: "Service is healthy",
	}

	if err != nil {
		apm.CaptureError(ctx, err).Send()
	}
	return response, err
}
