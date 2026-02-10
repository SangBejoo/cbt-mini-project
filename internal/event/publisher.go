package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// EventType defines the type of event
type EventType string

const (
	// Events published by CBT
	ExamResultCompleted EventType = "exam_result_completed"
)

// ExamResultPayload is sent when a student completes an exam
type ExamResultPayload struct {
	SessionID       int     `json:"session_id"`
	LMSAssignmentID int64   `json:"lms_assignment_id"`
	LMSUserID       int64   `json:"lms_user_id"`
	LMSClassID      int64   `json:"lms_class_id"`
	Score           float64 `json:"score"`
	CorrectCount    int     `json:"correct_count"`
	TotalCount      int     `json:"total_count"`
	CompletedAt     string  `json:"completed_at"`
}

// Publisher publishes events to Redis streams
type Publisher struct {
	client     *redis.Client
	streamName string
}

// NewPublisher creates a new event publisher
func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{
		client:     client,
		streamName: "cbt_events", // CBT publishes to this stream
	}
}

// Publish sends an event to the Redis stream
func (p *Publisher) Publish(ctx context.Context, eventType EventType, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"type":    string(eventType),
			"payload": string(data),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("[Publisher] Published %s event to %s", eventType, p.streamName)
	return nil
}

// PublishExamResult publishes an exam result completed event
func (p *Publisher) PublishExamResult(ctx context.Context, sessionID int, lmsAssignmentID, lmsUserID, lmsClassID int64, score float64, correctCount, totalCount int) error {
	payload := ExamResultPayload{
		SessionID:       sessionID,
		LMSAssignmentID: lmsAssignmentID,
		LMSUserID:       lmsUserID,
		LMSClassID:      lmsClassID,
		Score:           score,
		CorrectCount:    correctCount,
		TotalCount:      totalCount,
		CompletedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	return p.Publish(ctx, ExamResultCompleted, payload)
}
