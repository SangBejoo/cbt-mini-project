package config

import (
	"cbt-test-mini-project/util"

	"github.com/joho/godotenv"
)

type Main struct {
	Database   database
	Log        log
	RestServer restServer
	GrpcServer grpcServer
}

type database struct {
	DriverName string
	DSN        string
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

func Load() *Main {
	godotenv.Load()
	return &Main{
		Database: database{
			DriverName: util.GetEnv("DB_DRIVER", "mysql"),
			DSN:        util.GetEnv("DB_DSN", "root:root@tcp(localhost:3306)/cbt_test?charset=utf8mb4&parseTime=True&loc=Local"),
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
	}
}
