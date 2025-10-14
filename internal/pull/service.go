package pull

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"integration-education-db/internal/config"
	"integration-education-db/internal/db"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"

	"github.com/rs/zerolog"
)

type Service struct {
	cfg        *config.Config
	repo       db.Repository
	httpClient *http.Client
	log        zerolog.Logger
}

func NewService(cfg *config.Config, repo db.Repository) *Service {
	return &Service{
		cfg:  cfg,
		repo: repo,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		log: logger.Get(),
	}
}

// PullTeachers pulls teachers from server and upserts to local DB
func (s *Service) PullTeachers(ctx context.Context) (int, error) {
	s.log.Info().Msg("Pulling teachers from server")

	url := s.cfg.ExternalAPI.EducationDept.BaseURL + "/api/teachers"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch teachers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var teachers []model.ServerTeacher
	if err := json.NewDecoder(resp.Body).Decode(&teachers); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	s.log.Debug().Int("count", len(teachers)).Msg("Received teachers from server")

	// Upsert teachers to local database
	count := 0
	for _, teacher := range teachers {
		if err := s.repo.UpsertTeacher(ctx, &teacher); err != nil {
			s.log.Error().
				Err(err).
				Int64("teacher_id", teacher.ID).
				Str("name", teacher.Name).
				Msg("Failed to upsert teacher")
			continue
		}
		count++
	}

	s.log.Info().Int("success", count).Int("total", len(teachers)).Msg("Teachers sync completed")
	return count, nil
}

// PullStudents pulls students from server and upserts to local DB
func (s *Service) PullStudents(ctx context.Context) (int, error) {
	s.log.Info().Msg("Pulling students from server")

	url := s.cfg.ExternalAPI.EducationDept.BaseURL + "/api/students"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch students: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var students []model.ServerStudent
	if err := json.NewDecoder(resp.Body).Decode(&students); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	s.log.Debug().Int("count", len(students)).Msg("Received students from server")

	// Upsert students to local database
	count := 0
	for _, student := range students {
		if err := s.repo.UpsertStudent(ctx, &student); err != nil {
			s.log.Error().
				Err(err).
				Int64("student_id", student.ID).
				Str("name", student.Name).
				Msg("Failed to upsert student")
			continue
		}
		count++
	}

	s.log.Info().Int("success", count).Int("total", len(students)).Msg("Students sync completed")
	return count, nil
}

// PullClasses pulls classes from server and upserts to local DB
func (s *Service) PullClasses(ctx context.Context) (int, error) {
	s.log.Info().Msg("Pulling classes from server")

	url := s.cfg.ExternalAPI.EducationDept.BaseURL + "/api/classes"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch classes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var classes []model.ServerClass
	if err := json.NewDecoder(resp.Body).Decode(&classes); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	s.log.Debug().Int("count", len(classes)).Msg("Received classes from server")

	// Upsert classes to local database
	count := 0
	for _, class := range classes {
		if err := s.repo.UpsertClass(ctx, &class); err != nil {
			s.log.Error().
				Err(err).
				Int64("class_id", class.ID).
				Str("name", class.Name).
				Msg("Failed to upsert class")
			continue
		}
		count++
	}

	s.log.Info().Int("success", count).Int("total", len(classes)).Msg("Classes sync completed")
	return count, nil
}
