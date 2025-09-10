package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"integration-education-db/internal/config"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"
	"integration-education-db/pkg/errors"

	"github.com/rs/zerolog"
)

type Client struct {
	cfg         *config.Config
	httpClient  *http.Client
	authManager *AuthManager
	log         zerolog.Logger
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		authManager: NewAuthManager(cfg),
		log:         logger.Get(),
	}
}

func (c *Client) SendGradeBatch(ctx context.Context, grades []model.GradeStaging) (*model.BatchResponse, error) {
	if len(grades) == 0 {
		return nil, fmt.Errorf("empty grades batch")
	}

	// Convert to external format
	externalGrades := make([]model.ExternalGrade, len(grades))
	for i, grade := range grades {
		externalGrades[i] = model.ExternalGrade{
			StudentID: grade.StudentID,
			Subject:   grade.Subject,
			Class:     grade.Class,
			Semester:  grade.Semester,
			Grade:     grade.Grade,
		}
	}

	batch := model.GradeBatch{Grades: externalGrades}

	// Get auth token
	token, err := c.authManager.GetToken(ctx)
	if err != nil {
		return nil, errors.NewRetryableError(err, "failed to get auth token")
	}

	// Marshal request
	jsonData, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch: %w", err)
	}

	// Create request
	url := c.cfg.ExternalAPI.EducationDept.BaseURL + c.cfg.ExternalAPI.EducationDept.GradesEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	c.log.Debug().Int("batch_size", len(grades)).Msg("Sending grade batch to external API")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewRetryableError(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	// Handle response
	var batchResp model.BatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		c.log.Debug().Bool("success", batchResp.Success).Msg("Batch sent successfully")
		return &batchResp, nil
	case http.StatusUnauthorized:
		// Token might be expired, retry will refresh it
		return nil, errors.NewRetryableError(fmt.Errorf("unauthorized"), "authentication failed")
	case http.StatusBadRequest:
		// Business logic error - don't retry
		return &batchResp, nil
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		// Rate limited or service unavailable - retry
		return nil, errors.NewRetryableError(fmt.Errorf("service unavailable"), "external service unavailable")
	default:
		// Other errors - retry
		return nil, errors.NewRetryableError(fmt.Errorf("HTTP %d", resp.StatusCode), "external API error")
	}
}
