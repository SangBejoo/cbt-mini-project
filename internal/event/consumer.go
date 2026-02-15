package event

import (
	"context"
	"log/slog"

	infraRedis "cbt-test-mini-project/init/infra/redis"
	authRepo "cbt-test-mini-project/internal/repository/auth"
	classRepo "cbt-test-mini-project/internal/repository/class"
	classStudentRepo "cbt-test-mini-project/internal/repository/class_student"
	mataPelajaranRepo "cbt-test-mini-project/internal/repository/mata_pelajaran"
	"cbt-test-mini-project/internal/repository/materi"
	testSessionRepo "cbt-test-mini-project/internal/repository/test_session"
	"cbt-test-mini-project/internal/repository/tingkat"
	syncWorker "cbt-test-mini-project/internal/sync"
)

// Consumer is the LMS -> CBT event consumer entrypoint.
// It mirrors LMS's consumer bootstrap pattern while delegating
// domain-specific event processing to SyncWorker.
type Consumer struct {
	worker *syncWorker.SyncWorker
}

func NewConsumer(
	materiRepo materi.MateriRepository,
	tingkatRepo tingkat.TingkatRepository,
	subjectRepo mataPelajaranRepo.MataPelajaranRepository,
	authRepo authRepo.AuthRepository,
	testSessionRepo testSessionRepo.TestSessionRepository,
	classRepo classRepo.ClassRepository,
	classStudentRepo classStudentRepo.ClassStudentRepository,
) *Consumer {
	return &Consumer{
		worker: syncWorker.NewSyncWorker(
			materiRepo,
			tingkatRepo,
			subjectRepo,
			authRepo,
			testSessionRepo,
			classRepo,
			classStudentRepo,
		),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	if infraRedis.RedisClient == nil {
		slog.Warn("CBT event consumer disabled - Redis not available")
		return
	}

	if c.worker == nil {
		slog.Error("CBT event consumer disabled - worker is nil")
		return
	}

	slog.Info("CBT event consumer started", "stream", "lms_events")
	c.worker.Start(ctx)
}
