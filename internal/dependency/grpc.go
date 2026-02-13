package dependency

import (
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra"

	"google.golang.org/grpc"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/event"
	authHandler "cbt-test-mini-project/internal/handler/auth"
	baseGrpcServer "cbt-test-mini-project/internal/handler/base"
	historyHandler "cbt-test-mini-project/internal/handler/history"
	mataPelajaranHandler "cbt-test-mini-project/internal/handler/mata_pelajaran"
	materiHandler "cbt-test-mini-project/internal/handler/materi"
	soalHandler "cbt-test-mini-project/internal/handler/soal"
	soalDragDropHandler "cbt-test-mini-project/internal/handler/soal_drag_drop"
	testSessionHandler "cbt-test-mini-project/internal/handler/test_session"
	tingkatHandler "cbt-test-mini-project/internal/handler/tingkat"
	authRepo "cbt-test-mini-project/internal/repository/auth"
	historyRepo "cbt-test-mini-project/internal/repository/history"
	mataPelajaranRepo "cbt-test-mini-project/internal/repository/mata_pelajaran"
	materiRepo "cbt-test-mini-project/internal/repository/materi"
	soalDragDropRepo "cbt-test-mini-project/internal/repository/soal_drag_drop"
	testSessionRepo "cbt-test-mini-project/internal/repository/test_session"
	soalRepo "cbt-test-mini-project/internal/repository/test_soal"
	tingkatRepo "cbt-test-mini-project/internal/repository/tingkat"
	userLimitUsecase "cbt-test-mini-project/internal/usecase"
	authUsecase "cbt-test-mini-project/internal/usecase/auth"
	historyUsecase "cbt-test-mini-project/internal/usecase/history"
	mataPelajaranUsecase "cbt-test-mini-project/internal/usecase/mata_pelajaran"
	materiUsecase "cbt-test-mini-project/internal/usecase/materi"
	soalUsecase "cbt-test-mini-project/internal/usecase/soal"
	soalDragDropUsecase "cbt-test-mini-project/internal/usecase/soal_drag_drop"
	testSessionUsecase "cbt-test-mini-project/internal/usecase/test_session"
	tingkatUsecase "cbt-test-mini-project/internal/usecase/tingkat"
)

func InitGrpcDependency(server *grpc.Server, repo infra.Repository, config *config.Main, publisher *event.Publisher) {
	// Initialize repositories
	authRepo := authRepo.NewAuthRepository(repo.SQLDB)
	mataPelajaranRepo := mataPelajaranRepo.NewMataPelajaranRepository(repo.SQLDB)
	materiRepo := materiRepo.NewMateriRepository(repo.SQLDB)
	soalRepo := soalRepo.NewSoalRepository(repo.SQLDB)
	soalDragDropRepo := soalDragDropRepo.NewRepository(repo.SQLDB)
	testSessionRepo := testSessionRepo.NewTestSessionRepository(repo.SQLDB)
	historyRepo := historyRepo.NewHistoryRepository(repo.SQLDB)
	tingkatRepo := tingkatRepo.NewTingkatRepository(repo.SQLDB)

	// Initialize usecases
	authUsecase := authUsecase.NewAuthUsecase(authRepo, config)
	mataPelajaranUsecase := mataPelajaranUsecase.NewMataPelajaranUsecase(mataPelajaranRepo)
	materiUsecase := materiUsecase.NewMateriUsecase(materiRepo)
	soalUsecase := soalUsecase.NewSoalUsecase(soalRepo, config)
	soalDragDropUsecase := soalDragDropUsecase.NewUsecase(soalDragDropRepo, config)
	testSessionUsecase := testSessionUsecase.NewTestSessionUsecase(testSessionRepo, authRepo, publisher)
	historyUsecase := historyUsecase.NewHistoryUsecase(historyRepo)
	tingkatUsecase := tingkatUsecase.NewTingkatUsecase(tingkatRepo)
	userLimitUsecase := userLimitUsecase.NewUserLimitUsecase(repo.UserLimitRepo)

	// Initialize handlers
	baseServer := baseGrpcServer.NewBaseHandler()
	authServer := authHandler.NewAuthHandler(authUsecase)
	mataPelajaranServer := mataPelajaranHandler.NewMataPelajaranHandler(mataPelajaranUsecase)
	materiServer := materiHandler.NewMateriHandler(materiUsecase, soalUsecase, mataPelajaranUsecase)
	soalServer := soalHandler.NewSoalHandler(soalUsecase)
	soalDragDropServer := soalDragDropHandler.NewGrpcHandler(soalDragDropUsecase)
	testSessionServer := testSessionHandler.NewTestSessionHandler(testSessionUsecase, materiUsecase, tingkatUsecase, userLimitUsecase)
	historyServer := historyHandler.NewHistoryHandler(historyUsecase)
	tingkatServer := tingkatHandler.NewTingkatHandler(tingkatUsecase)

	// Register servers
	base.RegisterBaseServer(server, baseServer)
	base.RegisterAuthServiceServer(server, authServer)
	base.RegisterMataPelajaranServiceServer(server, mataPelajaranServer)
	base.RegisterMateriServiceServer(server, materiServer)
	base.RegisterSoalServiceServer(server, soalServer)
	base.RegisterSoalDragDropServiceServer(server, soalDragDropServer)
	base.RegisterTestSessionServiceServer(server, testSessionServer)
	base.RegisterHistoryServiceServer(server, historyServer)
	base.RegisterTingkatServiceServer(server, tingkatServer)
}