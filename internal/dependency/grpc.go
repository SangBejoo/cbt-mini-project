package dependency

import (
	"cbt-test-mini-project/init/infra"

	"google.golang.org/grpc"

	base "cbt-test-mini-project/gen/proto"
	baseGrpcServer "cbt-test-mini-project/internal/handler/base"
	historyHandler "cbt-test-mini-project/internal/handler/history"
	mataPelajaranHandler "cbt-test-mini-project/internal/handler/mata_pelajaran"
	materiHandler "cbt-test-mini-project/internal/handler/materi"
	soalHandler "cbt-test-mini-project/internal/handler/soal"
	testSessionHandler "cbt-test-mini-project/internal/handler/test_session"
	tingkatHandler "cbt-test-mini-project/internal/handler/tingkat"
	historyRepo "cbt-test-mini-project/internal/repository/history"
	mataPelajaranRepo "cbt-test-mini-project/internal/repository/mata_pelajaran"
	materiRepo "cbt-test-mini-project/internal/repository/materi"
	testSessionRepo "cbt-test-mini-project/internal/repository/test_session"
	soalRepo "cbt-test-mini-project/internal/repository/test_soal"
	tingkatRepo "cbt-test-mini-project/internal/repository/tingkat"
	historyUsecase "cbt-test-mini-project/internal/usecase/history"
	mataPelajaranUsecase "cbt-test-mini-project/internal/usecase/mata_pelajaran"
	materiUsecase "cbt-test-mini-project/internal/usecase/materi"
	soalUsecase "cbt-test-mini-project/internal/usecase/soal"
	testSessionUsecase "cbt-test-mini-project/internal/usecase/test_session"
	tingkatUsecase "cbt-test-mini-project/internal/usecase/tingkat"
)

func InitGrpcDependency(server *grpc.Server, repo infra.Repository) {
	// Initialize repositories
	mataPelajaranRepo := mataPelajaranRepo.NewMataPelajaranRepository(repo.GormDB)
	materiRepo := materiRepo.NewMateriRepository(repo.GormDB)
	soalRepo := soalRepo.NewSoalRepository(repo.GormDB)
	testSessionRepo := testSessionRepo.NewTestSessionRepository(repo.GormDB)
	historyRepo := historyRepo.NewHistoryRepository(repo.GormDB)
	tingkatRepo := tingkatRepo.NewTingkatRepository(repo.GormDB)

	// Initialize usecases
	mataPelajaranUsecase := mataPelajaranUsecase.NewMataPelajaranUsecase(mataPelajaranRepo)
	materiUsecase := materiUsecase.NewMateriUsecase(materiRepo)
	soalUsecase := soalUsecase.NewSoalUsecase(soalRepo)
	testSessionUsecase := testSessionUsecase.NewTestSessionUsecase(testSessionRepo)
	historyUsecase := historyUsecase.NewHistoryUsecase(historyRepo)
	tingkatUsecase := tingkatUsecase.NewTingkatUsecase(tingkatRepo)

	// Initialize handlers
	baseServer := baseGrpcServer.NewBaseHandler()
	mataPelajaranServer := mataPelajaranHandler.NewMataPelajaranHandler(mataPelajaranUsecase)
	materiServer := materiHandler.NewMateriHandler(materiUsecase)
	soalServer := soalHandler.NewSoalHandler(soalUsecase)
	testSessionServer := testSessionHandler.NewTestSessionHandler(testSessionUsecase, tingkatUsecase)
	historyServer := historyHandler.NewHistoryHandler(historyUsecase)
	tingkatServer := tingkatHandler.NewTingkatHandler(tingkatUsecase)

	// Register servers
	base.RegisterBaseServer(server, baseServer)
	base.RegisterMataPelajaranServiceServer(server, mataPelajaranServer)
	base.RegisterMateriServiceServer(server, materiServer)
	base.RegisterSoalServiceServer(server, soalServer)
	base.RegisterTestSessionServiceServer(server, testSessionServer)
	base.RegisterHistoryServiceServer(server, historyServer)
	base.RegisterTingkatServiceServer(server, tingkatServer)
}