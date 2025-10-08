package db

import (
	"context"
	"database/sql"

	"integration-education-db/internal/model"
)

type Repository interface {
	UpdateFileStatus(ctx context.Context, fileID int64, status model.FileStatus, errorMessage *string) error
	GetFile(ctx context.Context, fileID int64) (*model.File, error)
	InsertGrades(ctx context.Context, fileID int64, grades []model.GradeRow) error
	GetGradesByFileAndClassSubject(ctx context.Context, fileID int64, class, subject string) ([]model.GradeStaging, error)
	UpdateGradesStatus(ctx context.Context, ids []int64, status model.GradeStatus, errorMessage *string) error
	GetGradesStatus(ctx context.Context, fileID int64) (*model.StatusResponse, error)
	GetReadyGrades(ctx context.Context, fileID int64, class, subject string, limit int) ([]model.GradeStaging, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) UpdateFileStatus(ctx context.Context, fileID int64, status model.FileStatus, errorMessage *string) error {
	query := `UPDATE files SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, errorMessage, fileID)
	return err
}

func (r *repository) GetFile(ctx context.Context, fileID int64) (*model.File, error) {
	query := `SELECT id, s3_path, school_id, status, error_message, created_at, updated_at FROM files WHERE id = $1`

	var file model.File
	err := r.db.QueryRowContext(ctx, query, fileID).Scan(
		&file.ID, &file.S3Path, &file.SchoolID, &file.Status,
		&file.ErrorMessage, &file.CreatedAt, &file.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (r *repository) InsertGrades(ctx context.Context, fileID int64, grades []model.GradeRow) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO grades_staging (file_id, student_id, subject, class, semester, grade) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	for _, grade := range grades {
		_, err := tx.ExecContext(ctx, query, fileID, grade.StudentID, grade.Subject,
			grade.Class, grade.Semester, grade.Grade)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) GetGradesByFileAndClassSubject(ctx context.Context, fileID int64, class, subject string) ([]model.GradeStaging, error) {
	query := `SELECT id, file_id, student_id, subject, class, semester, grade, status, error_message, created_at, updated_at 
			  FROM grades_staging WHERE file_id = $1 AND class = $2 AND subject = $3`

	rows, err := r.db.QueryContext(ctx, query, fileID, class, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grades []model.GradeStaging
	for rows.Next() {
		var grade model.GradeStaging
		err := rows.Scan(&grade.ID, &grade.FileID, &grade.StudentID, &grade.Subject,
			&grade.Class, &grade.Semester, &grade.Grade, &grade.Status,
			&grade.ErrorMessage, &grade.CreatedAt, &grade.UpdatedAt)
		if err != nil {
			return nil, err
		}
		grades = append(grades, grade)
	}

	return grades, nil
}

func (r *repository) UpdateGradesStatus(ctx context.Context, ids []int64, status model.GradeStatus, errorMessage *string) error {
	if len(ids) == 0 {
		return nil
	}

	query := `UPDATE grades_staging SET status = $1, error_message = $2, updated_at = NOW() 
						  WHERE id = ANY($3)`

	_, err := r.db.ExecContext(ctx, query, status, errorMessage, ids)
	return err
}

func (r *repository) GetGradesStatus(ctx context.Context, fileID int64) (*model.StatusResponse, error) {
	query := `SELECT 
		COUNT(*) as total_records,
		COUNT(CASE WHEN status = 'SYNCED' THEN 1 END) as synced_count,
		COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed_count,
		MAX(updated_at) as updated_at
	FROM grades_staging WHERE file_id = $1`

	var response model.StatusResponse
	err := r.db.QueryRowContext(ctx, query, fileID).Scan(
		&response.TotalRecords, &response.SyncedCount,
		&response.FailedCount, &response.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	response.FileID = fileID

	// Get error messages for failed records
	errorQuery := `SELECT DISTINCT error_message FROM grades_staging 
				   WHERE file_id = $1 AND status = 'FAILED' AND error_message IS NOT NULL`

	rows, err := r.db.QueryContext(ctx, errorQuery, fileID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var errorMsg string
			if rows.Scan(&errorMsg) == nil {
				response.Errors = append(response.Errors, errorMsg)
			}
		}
	}

	return &response, nil
}

func (r *repository) GetReadyGrades(ctx context.Context, fileID int64, class, subject string, limit int) ([]model.GradeStaging, error) {
	query := `SELECT id, file_id, student_id, subject, class, semester, grade, status, error_message, created_at, updated_at 
			  FROM grades_staging 
			  WHERE file_id = $1 AND class = $2 AND subject = $3 AND status = 'READY' 
			  LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, fileID, class, subject, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grades []model.GradeStaging
	for rows.Next() {
		var grade model.GradeStaging
		err := rows.Scan(&grade.ID, &grade.FileID, &grade.StudentID, &grade.Subject,
			&grade.Class, &grade.Semester, &grade.Grade, &grade.Status,
			&grade.ErrorMessage, &grade.CreatedAt, &grade.UpdatedAt)
		if err != nil {
			return nil, err
		}
		grades = append(grades, grade)
	}

	return grades, nil
}
