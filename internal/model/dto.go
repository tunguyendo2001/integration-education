package model

import "time"

type IngestionJob struct {
	FileID   int64  `json:"file_id"`
	S3Path   string `json:"s3_path"`
	SchoolID int64  `json:"school_id"`
}

type SyncJob struct {
	FileID   int64  `json:"file_id"`
	Class    string `json:"class"`
	Subject  string `json:"subject"`
	Semester string `json:"semester"` // Added for permission check
	Year     int    `json:"year"`     // Added for permission check
}

type SyncRequest struct {
	FileID   int64  `json:"file_id"`
	Class    string `json:"class"`
	Subject  string `json:"subject"`
	Semester string `json:"semester"` // Added for permission check
	Year     int    `json:"year"`     // Added for permission check
}

type StatusResponse struct {
	FileID       int64     `json:"file_id"`
	Status       string    `json:"status"`
	TotalRecords int       `json:"total_records"`
	SyncedCount  int       `json:"synced_count"`
	FailedCount  int       `json:"failed_count"`
	Errors       []string  `json:"errors,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AuthTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type GradeBatch struct {
	Grades []ExternalGrade `json:"grades"`
}

type ExternalGrade struct {
	StudentID string  `json:"student_id"`
	Subject   string  `json:"subject"`
	Class     string  `json:"class"`
	Semester  string  `json:"semester"`
	Grade     float64 `json:"grade"`
}

type BatchResponse struct {
	Success bool     `json:"success"`
	Failed  []string `json:"failed,omitempty"`
	Message string   `json:"message,omitempty"`
}
