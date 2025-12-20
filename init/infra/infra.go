package infra

import (
	"log"
	"os"

	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra/db"

	"go.elastic.co/apm"
	"gorm.io/gorm"
)

type Repository struct {
	GormDB *gorm.DB
}

func (r *Repository) Close() error {
	if r != nil && r.GormDB != nil {
		sqlDB, _ := r.GormDB.DB()
		if err := sqlDB.Close(); err != nil {
			return err
		}
	}
	return nil
}

func LoadRepository(cfg config.Main) *Repository {
	dbConn, err := db.OpenSQL(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✓ Database connected successfully")

	return &Repository{
		GormDB: dbConn,
	}
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
