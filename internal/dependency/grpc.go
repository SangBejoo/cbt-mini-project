package dependency

import (
	"cbt-test-mini-project/init/infra"

	"google.golang.org/grpc"

	base "cbt-test-mini-project/gen/proto"
	baseGrpcServer "cbt-test-mini-project/internal/handler/base"
	notesGrpcHandler "cbt-test-mini-project/internal/handler/notes"
	NotesRepository "cbt-test-mini-project/internal/repository/notes"
	NotesUseCase "cbt-test-mini-project/internal/usecase/notes"
)

func InitGrpcDependency(server *grpc.Server, repo infra.Repository) {
	baseServer := baseGrpcServer.NewBaseHandler()
	base.RegisterBaseServer(server, baseServer)
	notesRepository := NotesRepository.NewNotesRepository(repo.DB)
	notesUseCase := NotesUseCase.NewNotesUseCase(notesRepository)
	notesServer := notesGrpcHandler.NewNotesHandler(notesUseCase)
	base.RegisterNotesServiceServer(server, notesServer)
}
