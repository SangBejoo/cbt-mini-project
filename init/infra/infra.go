package infra

import (
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra/db"
	"database/sql"
	"log"
	"os"

	"go.elastic.co/apm"

	"cbt-test-mini-project/internal/repository"
)

type Repository struct {
	SQLDB         *sql.DB
	UserLimitRepo repository.UserLimitRepository
}

func (r *Repository) Close() error {
	if r != nil && r.SQLDB != nil {
		if err := r.SQLDB.Close(); err != nil {
			return err
		}
	}
	return nil
}

func LoadRepository(cfg config.Main) *Repository {
	sqlDB, err := db.OpenSQL(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✓ Database connected successfully")

	repo := &Repository{
		SQLDB: sqlDB,
	}

	// Initialize user limit repository
	repo.UserLimitRepo = repository.NewUserLimitRepository(sqlDB, &cfg)

	return repo
}

func InitAPM(cfg config.Main) {
	if !cfg.APM.Enabled {
		log.Println("APM is disabled")
		return
	}

	// Set environment variables for APM configuration
	os.Setenv("ELASTIC_APM_SERVER_URL", cfg.APM.ServerURL)
	os.Setenv("ELASTIC_APM_SERVICE_NAME", cfg.APM.ServiceName)
	os.Setenv("ELASTIC_APM_SERVICE_VERSION", cfg.APM.ServiceVersion)
	os.Setenv("ELASTIC_APM_ENVIRONMENT", cfg.APM.Environment)

	// Initialize Elastic APM with environment variables
	tracer, err := apm.NewTracer(cfg.APM.ServiceName, cfg.APM.ServiceVersion)
	if err != nil {
		log.Printf("Failed to initialize APM tracer: %v", err)
		return
	}

	apm.DefaultTracer = tracer

	log.Printf("✓ APM initialized: service=%s, version=%s, environment=%s, server=%s",
		cfg.APM.ServiceName, cfg.APM.ServiceVersion, cfg.APM.Environment, cfg.APM.ServerURL)
}
