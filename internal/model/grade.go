package model

import "time"

type GradeStatus string

const (
	GradeStatusReady  GradeStatus = "READY"
	GradeStatusSynced GradeStatus = "SYNCED"
	GradeStatusFailed GradeStatus = "FAILED"
)

type GradeStaging struct {
	ID           int64       `json:"id" db:"id"`
	FileID       int64       `json:"file_id" db:"file_id"`
	StudentID    string      `json:"student_id" db:"student_id"`
	Subject      string      `json:"subject" db:"subject"`
	Class        string      `json:"class" db:"class"`
	Semester     string      `json:"semester" db:"semester"`
	Grade        float64     `json:"grade" db:"grade"`
	Status       GradeStatus `json:"status" db:"status"`
	ErrorMessage *string     `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

type GradeRow struct {
	StudentID string  `json:"student_id"`
	Subject   string  `json:"subject"`
	Class     string  `json:"class"`
	Semester  string  `json:"semester"`
	Grade     float64 `json:"grade"`
}
