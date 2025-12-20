# CBT Mini Project

Computer-Based Test (CBT) system for educational institutions with comprehensive monitoring and analytics.

## Tech Stack

- **Backend**: Go 1.21+, gRPC, REST Gateway
- **Database**: MySQL with GORM
- **Monitoring**: Elastic APM (Elasticsearch, Kibana)
- **Authentication**: JWT
- **Frontend**: Next.js, TypeScript
- **Deployment**: Docker, Docker Compose

## Features

### Core Features
- ✅ User authentication (Admin, Student)
- ✅ Test session management
- ✅ Question management (Multiple choice, Essay)
- ✅ Answer submission and grading
- ✅ Student history tracking
- ✅ Subject and level management

### Monitoring & Analytics
- ✅ Real-time performance monitoring
- ✅ Transaction tracing (gRPC/HTTP)
- ✅ Error tracking
- ✅ Database query monitoring
- ✅ Service health checks

## Prerequisites

- Go 1.21+
- MySQL 8.0+
- Node.js 18+
- Docker & Docker Compose

## Quick Start

### 1. Clone & Setup
```bash
git clone <repository-url>
cd cbt-mini-project
cp .env.example .env
```

### 2. Start Infrastructure
```bash
cd deployment
docker-compose up -d
```

### 3. Run Backend
```bash
go run main.go
```

### 4. Run Frontend
```bash
cd web
npm install
npm run dev
```

## API Endpoints

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout

### Test Management
- `GET /api/test-sessions` - List test sessions
- `POST /api/test-sessions` - Create test session
- `GET /api/test-sessions/{id}/questions` - Get test questions
- `POST /api/test-sessions/{id}/submit` - Submit answers

### Admin Features
- `GET /api/users` - List users
- `POST /api/questions` - Create questions
- `GET /api/subjects` - List subjects
- `GET /api/history` - Student history

## Monitoring Dashboard

- **Kibana**: http://localhost:5601
- **APM Server**: http://localhost:8200
- **Elasticsearch**: http://localhost:9200

## Development

### Project Structure
```
├── main.go                 # Application entry point
├── init/                   # Initialization modules
│   ├── config/            # Configuration management
│   ├── infra/             # Infrastructure setup
│   ├── logger/            # Logging setup
│   └── server/            # Server setup
├── internal/               # Business logic
│   ├── entity/            # Data models
│   ├── handler/           # HTTP handlers
│   ├── repository/        # Data access layer
│   └── usecase/           # Business logic layer
├── util/                   # Utilities
├── web/                    # Frontend application
├── deployment/             # Docker deployment
└── databases/              # Database migrations
```

### Environment Variables

```env
# Database
DB_DSN=root:root@tcp(localhost:3306)/cbt_test

# Server
GRPC_PORT=6000
REST_PORT=8080

# JWT
JWT_SECRET=your-secret-key

# APM
ELASTIC_APM_SERVER_URL=http://localhost:8200
ELASTIC_APM_SERVICE_NAME=cbt-mini-project
```

## License

MIT License