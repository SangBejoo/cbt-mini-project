package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Update this import path
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra"
	infraRedis "cbt-test-mini-project/init/infra/redis"
	"cbt-test-mini-project/internal/dependency"
	"cbt-test-mini-project/internal/event"
	classRepo "cbt-test-mini-project/internal/repository/class"
	classStudentRepo "cbt-test-mini-project/internal/repository/class_student"
)

// ShareEmailRequest represents the request payload for sharing results via email
type ShareEmailRequest struct {
	To              string `json:"to"`
	Subject         string `json:"subject"`
	NamaSekolah     string `json:"namaSekolah"`
	Kelas           string `json:"kelas"`
	StudentName     string `json:"studentName"`
	SubjectName     string `json:"subject_name"`
	LevelName       string `json:"level_name"`
	Score           float64 `json:"score"`
	CorrectAnswers  int32 `json:"correctAnswers"`
	TotalQuestions  int32 `json:"totalQuestions"`
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	Duration        int32 `json:"duration"`
	Status          string `json:"status"`
	SessionToken    string `json:"sessionToken"`
}

// ShareEmailResponse represents the response for email sharing
type ShareEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SyncOpsHandler struct {
	classRepo        classRepo.ClassRepository
	classStudentRepo classStudentRepo.ClassStudentRepository
	db               *sql.DB
}

func NewSyncOpsHandler(db *sql.DB) *SyncOpsHandler {
	return &SyncOpsHandler{
		classRepo:        classRepo.NewClassRepository(db),
		classStudentRepo: classStudentRepo.NewClassStudentRepository(db),
		db:               db,
	}
}

type SyncHealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Redis     string    `json:"redis"`
	Timestamp time.Time `json:"timestamp"`
}

type SyncClassDTO struct {
	LMSClassID  int64  `json:"lms_class_id"`
	LMSSchoolID int64  `json:"lms_school_id"`
	Name        string `json:"name"`
	IsActive    bool   `json:"is_active"`
}

type SyncClassStudentDTO struct {
	LMSClassID int64     `json:"lms_class_id"`
	LMSUserID  int64     `json:"lms_user_id"`
	JoinedAt   time.Time `json:"joined_at"`
}

type SyncClassesResponse struct {
	Data  []SyncClassDTO `json:"data"`
	Total int            `json:"total"`
}

type SyncClassStudentsResponse struct {
	LMSClassID int64                `json:"lms_class_id"`
	Data       []SyncClassStudentDTO `json:"data"`
	Total      int                  `json:"total"`
}

func RunGatewayRestServer(ctx context.Context, cfg config.Main, repo infra.Repository, publisher *event.Publisher) (*http.Server, error) {
	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(customHeaderMatcher),
		runtime.WithErrorHandler(customErrorHandler),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Register your services here
	dependency.InitRestGatewayDependency(gwMux, opts, ctx, cfg, publisher)

	// Create a custom mux to handle both API and static files
	mux := http.NewServeMux()
	syncOpsHandler := NewSyncOpsHandler(repo.SQLDB)

	// Custom endpoints
	mux.HandleFunc("/v1/sessions/share-email", handleShareEmail)
	mux.HandleFunc("/v1/sync/health", syncOpsHandler.HandleSyncHealth)
	mux.HandleFunc("/v1/sync/classes", syncOpsHandler.HandleSyncClasses)
	mux.HandleFunc("/v1/sync/classes/", syncOpsHandler.HandleSyncClassStudents)

	// Serve static files (uploads)
	fs := http.FileServer(http.Dir("uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	// Serve API through gRPC-Gateway
	mux.Handle("/", gwMux)

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.RestServer.Port),
		Handler:      corsMiddleware(&cfg)(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("starting REST gateway server", "port", cfg.RestServer.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start REST gateway server", "error", err)
		}
	}()

	return srv, nil
}

// handleShareEmail handles email sharing requests
func handleShareEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ShareEmailResponse{
			Success: false,
			Message: "Failed to read request body",
		})
		return
	}
	defer r.Body.Close()

	var req ShareEmailRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		slog.Error("Failed to unmarshal request", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ShareEmailResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate required fields
	if req.To == "" || req.StudentName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ShareEmailResponse{
			Success: false,
			Message: "Email and student name are required",
		})
		return
	}

	// For now, just simulate successful email sending
	// In production, integrate with a real email service (SendGrid, Mailgun, etc.)
	slog.Info("Email sharing request received",
		"to", req.To,
		"student", req.StudentName,
		"subject", req.SubjectName,
		"score", req.Score,
	)

	// Simulate email sending
	success := sendEmailNotification(req)

	if success {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ShareEmailResponse{
			Success: true,
			Message: fmt.Sprintf("Email successfully sent to %s", req.To),
		})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ShareEmailResponse{
			Success: false,
			Message: "Failed to send email",
		})
	}
}

func (h *SyncOpsHandler) HandleSyncHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	dbStatus := "ok"
	if err := h.db.PingContext(r.Context()); err != nil {
		dbStatus = "error"
	}

	redisStatus := "disconnected"
	if infraRedis.RedisClient != nil {
		if _, err := infraRedis.RedisClient.Ping(r.Context()).Result(); err == nil {
			redisStatus = "ok"
		} else {
			redisStatus = "error"
		}
	}

	status := "ok"
	if dbStatus != "ok" || redisStatus == "error" {
		status = "degraded"
	}

	_ = json.NewEncoder(w).Encode(SyncHealthResponse{
		Status:    status,
		Database:  dbStatus,
		Redis:     redisStatus,
		Timestamp: time.Now(),
	})
}

func (h *SyncOpsHandler) HandleSyncClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	classes, err := h.classRepo.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to list classes"})
		return
	}

	result := make([]SyncClassDTO, 0, len(classes))
	for _, item := range classes {
		result = append(result, SyncClassDTO{
			LMSClassID:  item.LMSClassID,
			LMSSchoolID: item.LMSSchoolID,
			Name:        item.Name,
			IsActive:    item.IsActive,
		})
	}

	_ = json.NewEncoder(w).Encode(SyncClassesResponse{Data: result, Total: len(result)})
}

func (h *SyncOpsHandler) HandleSyncClassStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/v1/sync/classes/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "students" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path, expected /v1/sync/classes/{lms_class_id}/students"})
		return
	}

	lmsClassID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || lmsClassID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid lms_class_id"})
		return
	}

	students, err := h.classStudentRepo.ListByClassID(lmsClassID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to list class students"})
		return
	}

	result := make([]SyncClassStudentDTO, 0, len(students))
	for _, item := range students {
		result = append(result, SyncClassStudentDTO{
			LMSClassID: item.LMSClassID,
			LMSUserID:  item.LMSUserID,
			JoinedAt:   item.JoinedAt,
		})
	}

	_ = json.NewEncoder(w).Encode(SyncClassStudentsResponse{LMSClassID: lmsClassID, Data: result, Total: len(result)})
}

// sendEmailNotification sends an email notification
func sendEmailNotification(req ShareEmailRequest) bool {
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")

	if from == "" || password == "" || smtpServer == "" || smtpPort == "" {
		slog.Error("Email configuration missing", "from", from, "server", smtpServer, "port", smtpPort)
		return false
	}

	to := []string{req.To}
	subject := fmt.Sprintf("Hasil Tes CBT - %s", req.StudentName)
	body := buildEmailBody(req)
	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	auth := smtp.PlainAuth("", from, password, smtpServer)
	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, from, to, []byte(message))
	if err != nil {
		slog.Error("Failed to send email", "error", err, "to", req.To)
		return false
	}

	slog.Info("Email notification sent",
		"to", req.To,
		"student", req.StudentName,
		"school", req.NamaSekolah,
		"class", req.Kelas,
		"score", req.Score,
	)
	return true
}

// Custom header matcher for gRPC-Gateway
func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "authorization":
		return key, true
	case "x-request-id":
		return key, true
	default:
		return key, false
	}
}

// Custom error handler for gRPC-Gateway
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
	// Log the actual error with full details
	slog.Error("=== gRPC Gateway Error ===",
		"error", err,
		"path", req.URL.Path,
		"method", req.Method,
		"remote_addr", req.RemoteAddr,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	response := map[string]interface{}{
		"error":   "Internal Server Error",
		"message": "Something went wrong",
		"code":    500,
		"details": err.Error(), // Add actual error message for debugging
	}

	if jsonErr := json.NewEncoder(w).Encode(response); jsonErr != nil {
		slog.Error("Failed to encode error response", "error", jsonErr)
	}
}

// buildEmailBody builds the email body
func buildEmailBody(req ShareEmailRequest) string {
	return fmt.Sprintf(`Halo,

Berikut adalah hasil tes CBT untuk siswa:

Nama Siswa: %s
Sekolah: %s
Kelas: %s
Mata Pelajaran: %s
Tingkat: %s
Nilai Akhir: %.2f%%
Jumlah Benar: %d/%d
Waktu Mulai: %s
Waktu Selesai: %s
Durasi: %d menit
Status: %s

Terima kasih telah menggunakan sistem CBT.

Salam,
Tim CBT
`,
		req.StudentName,
		req.NamaSekolah,
		req.Kelas,
		req.SubjectName,
		req.LevelName,
		req.Score,
		req.CorrectAnswers,
		req.TotalQuestions,
		req.StartTime,
		req.EndTime,
		req.Duration,
		strings.Title(req.Status),
	)
}
