package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Update this import path
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra"
	"cbt-test-mini-project/internal/dependency"
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

func RunGatewayRestServer(ctx context.Context, cfg config.Main, repo infra.Repository) (*http.Server, error) {
	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(customHeaderMatcher),
		runtime.WithErrorHandler(customErrorHandler),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Register your services here
	dependency.InitRestGatewayDependency(gwMux, opts, ctx, cfg)

	// Create a custom mux to handle both API and static files
	mux := http.NewServeMux()

	// Serve static files from uploads directory
	uploadsDir := "./uploads"
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// Custom endpoints
	mux.HandleFunc("/v1/sessions/share-email", handleShareEmail)

	// Serve API through gRPC-Gateway
	mux.Handle("/", gwMux)

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.RestServer.Port),
		Handler:      corsMiddleware(mux),
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	response := map[string]interface{}{
		"error":   "Internal Server Error",
		"message": "Something went wrong",
		"code":    500,
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
