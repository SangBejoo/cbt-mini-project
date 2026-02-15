package user_limit

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	userLimitUsecase "cbt-test-mini-project/internal/usecase"
	"cbt-test-mini-project/util/interceptor"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userLimitHandler struct {
	base.UnimplementedUserLimitServiceServer
	usecase userLimitUsecase.UserLimitUsecase
}

func NewUserLimitHandler(usecase userLimitUsecase.UserLimitUsecase) base.UserLimitServiceServer {
	return &userLimitHandler{usecase: usecase}
}

func (h *userLimitHandler) requireAdmin(ctx context.Context) error {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if user.Role != base.UserRole_ADMIN {
		return status.Error(codes.PermissionDenied, "only admin can access this endpoint")
	}
	return nil
}

func (h *userLimitHandler) GetUserLimits(ctx context.Context, req *base.GetUserLimitsRequest) (*base.GetUserLimitsResponse, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id must be greater than 0")
	}

	limits, err := h.usecase.GetUserLimits(ctx, int(req.UserId))
	if err != nil {
		return &base.GetUserLimitsResponse{Success: false, Message: err.Error()}, nil
	}

	res := make([]*base.UserLimit, 0, len(limits))
	for _, limit := range limits {
		res = append(res, convertUserLimit(limit))
	}

	return &base.GetUserLimitsResponse{
		Limits:  res,
		Success: true,
		Message: "User limits retrieved successfully",
	}, nil
}

func (h *userLimitHandler) SetUserLimit(ctx context.Context, req *base.SetUserLimitRequest) (*base.UserLimitResponse, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id must be greater than 0")
	}
	if req.LimitType == "" {
		return nil, status.Error(codes.InvalidArgument, "limit_type is required")
	}
	if req.LimitValue < 0 {
		return nil, status.Error(codes.InvalidArgument, "limit_value cannot be negative")
	}

	limit, err := h.usecase.SetUserLimit(ctx, int(req.UserId), req.LimitType, int(req.LimitValue))
	if err != nil {
		return &base.UserLimitResponse{Success: false, Message: err.Error()}, nil
	}

	return &base.UserLimitResponse{
		Limit:   convertUserLimit(limit),
		Success: true,
		Message: "User limit updated successfully",
	}, nil
}

func (h *userLimitHandler) ResetUserLimit(ctx context.Context, req *base.ResetUserLimitRequest) (*base.MessageStatusResponse, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id must be greater than 0")
	}
	if req.LimitType == "" {
		return nil, status.Error(codes.InvalidArgument, "limit_type is required")
	}

	if err := h.usecase.ResetLimit(ctx, int(req.UserId), req.LimitType); err != nil {
		return &base.MessageStatusResponse{Status: "error", Message: err.Error()}, nil
	}

	return &base.MessageStatusResponse{Status: "success", Message: "User limit reset successfully"}, nil
}

func (h *userLimitHandler) GetUserLimitUsageHistory(ctx context.Context, req *base.GetUserLimitUsageHistoryRequest) (*base.GetUserLimitUsageHistoryResponse, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id must be greater than 0")
	}
	if req.LimitType == "" {
		return nil, status.Error(codes.InvalidArgument, "limit_type is required")
	}
	days := int(req.Days)
	if days <= 0 {
		days = 7
	}

	history, err := h.usecase.GetUsageHistory(ctx, int(req.UserId), req.LimitType, days)
	if err != nil {
		return &base.GetUserLimitUsageHistoryResponse{Success: false, Message: err.Error()}, nil
	}

	res := make([]*base.UserLimitUsage, 0, len(history))
	for _, item := range history {
		protoItem := &base.UserLimitUsage{
			Id:        int32(item.ID),
			UserId:    int32(item.UserID),
			LimitType: item.LimitType,
			Action:    item.Action,
			CreatedAt: timestamppb.New(item.CreatedAt),
		}
		if item.ResourceID != nil {
			protoItem.ResourceId = int32(*item.ResourceID)
			protoItem.HasResourceId = true
		}
		res = append(res, protoItem)
	}

	return &base.GetUserLimitUsageHistoryResponse{
		History: res,
		Success: true,
		Message: "User limit usage history retrieved successfully",
	}, nil
}

func convertUserLimit(limit *entity.UserLimit) *base.UserLimit {
	if limit == nil {
		return nil
	}
	return &base.UserLimit{
		Id:          int32(limit.ID),
		UserId:      int32(limit.UserID),
		LimitType:   limit.LimitType,
		LimitValue:  int32(limit.LimitValue),
		CurrentUsed: int32(limit.CurrentUsed),
		ResetAt:     timestamppb.New(limit.ResetAt),
		CreatedAt:   timestamppb.New(limit.CreatedAt),
		UpdatedAt:   timestamppb.New(limit.UpdatedAt),
	}
}
