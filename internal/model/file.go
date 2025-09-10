package model

import "time"

type FileStatus string

const (
	FileStatusUploaded   FileStatus = "UPLOADED"
	FileStatusParsedOK   FileStatus = "PARSED_OK"
	FileStatusParsedFail FileStatus = "PARSED_FAIL"
)

type File struct {
	ID           int64      `json:"id" db:"id"`
	S3Path       string     `json:"s3_path" db:"s3_path"`
	SchoolID     int64      `json:"school_id" db:"school_id"`
	Status       FileStatus `json:"status" db:"status"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}
