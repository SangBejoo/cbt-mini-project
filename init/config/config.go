package config

import (
	"cbt-test-mini-project/util"

	"github.com/joho/godotenv"
)

type Main struct {
	Database   Database
	Log        log
	RestServer restServer
	GrpcServer grpcServer
	JWT        jwt
	APM        apm
}

type Database struct {
	DriverName       string
	DSN              string
	MaxOpenConns     int
	MaxIdleConns     int
	MinIdleConns     int // Minimum idle connections to maintain
	ConnMaxLifetime  int // in minutes
}

type log struct {
	Level     int
	Directory string
}

type restServer struct {
	Port int
}

type grpcServer struct {
	Port int
}

type jwt struct {
	Secret           string
	AccessTokenTTL   int // in minutes
	RefreshTokenTTL  int // in days
}

type apm struct {
	ServerURL      string
	ServiceName    string
	ServiceVersion string
	Environment    string
	Enabled        bool
}

func Load() *Main {
	godotenv.Load()
	return &Main{
		Database: Database{
			DriverName:     util.GetEnv("DB_DRIVER", "mysql"),
			DSN:            util.GetEnv("DB_DSN", "root:root@tcp(localhost:3306)/cbt_test?charset=utf8mb4&parseTime=True&loc=Local"),
			MaxOpenConns:   util.GetEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:   util.GetEnv("DB_MAX_IDLE_CONNS", 25),
			MinIdleConns:   util.GetEnv("DB_MIN_IDLE_CONNS", 5),
			ConnMaxLifetime: util.GetEnv("DB_CONN_MAX_LIFETIME_MINUTES", 5),
		},
		Log: log{
			Level:     util.GetEnv("LOG_LEVEL", -1),
			Directory: util.GetEnv("LOG_DIRECTORY", ""),
		},
		RestServer: restServer{
			Port: util.GetEnv("REST_PORT", 8080),
		},
		GrpcServer: grpcServer{
			Port: util.GetEnv("GRPC_PORT", 6000),
		},
		JWT: jwt{
			Secret:          util.GetEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
			AccessTokenTTL:  util.GetEnv("JWT_ACCESS_TTL_MINUTES", 120), // 2 hours default
			RefreshTokenTTL: util.GetEnv("JWT_REFRESH_TTL_MINUTES", 240),     // 4 hours default
		},
		APM: apm{
			ServerURL:      util.GetEnv("ELASTIC_APM_SERVER_URL", "http://localhost:8200"),
			ServiceName:    util.GetEnv("ELASTIC_APM_SERVICE_NAME", "cbt-mini-project"),
			ServiceVersion: util.GetEnv("ELASTIC_APM_SERVICE_VERSION", "1.0.0"),
			Environment:    util.GetEnv("ELASTIC_APM_ENVIRONMENT", "development"),
			Enabled:        util.GetEnv("ELASTIC_APM_ENABLED", true),
		},
	}
}
