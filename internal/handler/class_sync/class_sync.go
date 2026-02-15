package class_sync

import (
	base "cbt-test-mini-project/gen/proto"
	classUsecase "cbt-test-mini-project/internal/usecase/class"
	classStudentUsecase "cbt-test-mini-project/internal/usecase/class_student"
	"cbt-test-mini-project/util/interceptor"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type classSyncHandler struct {
	base.UnimplementedClassSyncServiceServer
	classUsecase        classUsecase.ClassUsecase
	classStudentUsecase classStudentUsecase.ClassStudentUsecase
}

func NewClassSyncHandler(classUsecase classUsecase.ClassUsecase, classStudentUsecase classStudentUsecase.ClassStudentUsecase) base.ClassSyncServiceServer {
	return &classSyncHandler{
		classUsecase:        classUsecase,
		classStudentUsecase: classStudentUsecase,
	}
}

func (h *classSyncHandler) ListClasses(ctx context.Context, req *base.ListClassesRequest) (*base.ListClassesResponse, error) {
	if err := h.ensureAdmin(ctx); err != nil {
		return nil, err
	}

	classes, err := h.classUsecase.ListClasses(req.LmsSchoolId)
	if err != nil {
		return nil, err
	}

	result := make([]*base.ClassData, 0, len(classes))
	for _, item := range classes {
		result = append(result, &base.ClassData{
			Id:          int32(item.ID),
			LmsClassId:  item.LMSClassID,
			LmsSchoolId: item.LMSSchoolID,
			Name:        item.Name,
			IsActive:    item.IsActive,
			CreatedAt:   timestamppb.New(item.CreatedAt),
			UpdatedAt:   timestamppb.New(item.UpdatedAt),
		})
	}

	return &base.ListClassesResponse{Classes: result}, nil
}

func (h *classSyncHandler) ListClassStudents(ctx context.Context, req *base.ListClassStudentsRequest) (*base.ListClassStudentsResponse, error) {
	if err := h.ensureAdmin(ctx); err != nil {
		return nil, err
	}

	students, err := h.classStudentUsecase.ListClassStudents(req.LmsClassId)
	if err != nil {
		return nil, err
	}

	result := make([]*base.ClassStudentData, 0, len(students))
	for _, item := range students {
		result = append(result, &base.ClassStudentData{
			Id:         int32(item.ID),
			LmsClassId: item.LMSClassID,
			LmsUserId:  item.LMSUserID,
			JoinedAt:   timestamppb.New(item.JoinedAt),
		})
	}

	return &base.ListClassStudentsResponse{Students: result}, nil
}

func (h *classSyncHandler) ensureAdmin(ctx context.Context) error {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if user.Role != base.UserRole_ADMIN {
		return status.Error(codes.PermissionDenied, "admin role required")
	}

	return nil
}
