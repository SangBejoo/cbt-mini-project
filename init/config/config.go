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
}

type Database struct {
	DriverName     string
	DSN            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime int // in minutes
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

func Load() *Main {
	godotenv.Load()
	return &Main{
		Database: Database{
			DriverName:     util.GetEnv("DB_DRIVER", "mysql"),
			DSN:            util.GetEnv("DB_DSN", "root:root@tcp(localhost:3306)/cbt_test?charset=utf8mb4&parseTime=True&loc=Local"),
			MaxOpenConns:   util.GetEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:   util.GetEnv("DB_MAX_IDLE_CONNS", 25),
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
	}
}
