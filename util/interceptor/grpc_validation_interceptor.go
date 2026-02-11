package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GRPCValidationInterceptor returns a gRPC unary interceptor for validation
func GRPCValidationInterceptor() grpc.UnaryServerInterceptor {
	validator := NewValidationInterceptor()

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip validation for certain methods
		if shouldSkipValidation(info.FullMethod) {
			return handler(ctx, req)
		}

		// Convert request to protoreflect.Message if possible
		if protoMsg, ok := req.(protoreflect.ProtoMessage); ok {
			if err := validator.Do(ctx, protoMsg); err != nil {
				// Return validation error as gRPC status
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		// Continue with handler
		return handler(ctx, req)
	}
}

// shouldSkipValidation determines if a gRPC method should skip validation
func shouldSkipValidation(fullMethod string) bool {
	// Skip health checks
	if strings.Contains(fullMethod, "Health") {
		return true
	}

	// Skip read-only operations (optional - you can enable validation for query params)
	readOnlyMethods := []string{"Get", "List", "Search"}
	for _, method := range readOnlyMethods {
		if strings.Contains(fullMethod, "/"+method) {
			return true
		}
	}

	return false
}