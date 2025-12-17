# Default service name
SERVICE_NAME := template

# Override SERVICE_NAME if set in environment
ifneq ($(strip $(SERVICE_NAME)),)
    SERVICE_NAME := $(SERVICE_NAME)
endif

start: 
	docker compose up -d --no-deps --build app migrations

proto:
	cd contract && \
	protoc \
	--go_out=../gen/proto --go_opt=paths=source_relative \
	--go-grpc_out=../gen/proto --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=../gen/proto \
	--grpc-gateway_opt=paths=source_relative \
	--grpc-gateway_opt=grpc_api_configuration=gateway_template.yaml \
	--openapiv2_out=../gen/swagger \
	--openapiv2_opt=allow_merge=true,merge_file_name=api \
	--openapiv2_opt=grpc_api_configuration=gateway_template.yaml \
	template.proto

# Docker build target
docker-build:
	@if [ -z "$$GIT_USER" ]; then \
		echo "Error: GIT_USER environment variable is not set."; \
		echo "Usage: GIT_USER=your_git_username GIT_TOKEN=your_git_token make docker-build"; \
		exit 1; \
	fi
	@if [ -z "$$GIT_TOKEN" ]; then \
		echo "Error: GIT_TOKEN environment variable is not set."; \
		echo "Usage: GIT_USER=your_git_username GIT_TOKEN=your_git_token make docker-build"; \
		exit 1; \
	fi
	docker build \
		--build-arg GIT_USER=$$GIT_USER \
		--build-arg GIT_TOKEN=$$GIT_TOKEN \
		-t $(SERVICE_NAME) \
		-f deployment/Dockerfile .

docker-run:
	docker run -d -p 6000:6000 -p 8080:8080 -p 6015:6015 $(SERVICE_NAME)
# Database connection parameters
DB_USER ?= user
DB_PASSWORD ?= password
DB_NAME ?= testdb
DB_PORT ?= 5432
migrate-up:
	cd deployment && docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	cd deployment && docker compose exec -T db psql -U $(DB_USER) -d $(DB_NAME) -f /docker-entrypoint-initdb.d/data.sql
migrate-down:
	@if [ -z "$$DB_URL" ]; then \
		echo "Error: DB_URL environment variable is not set."; \
		echo "Usage: DB_URL=your_database_url make migrate-down"; \
		exit 1; \
	fi
	dbmate --no-dump-schema -d migration/db/script -u $$DB_URL down

migration-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: Please provide a name for the migration."; \
		echo "Usage: make migration-new name=your_migration_name"; \
		exit 1; \
	fi
	dbmate --no-dump-schema -d migration/db/script new $(name)

seed:
	psql $(DB_URL) < migration/db/seed/init_seed.sql