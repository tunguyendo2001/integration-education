package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"integration-education-db/internal/config"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"

	"github.com/rs/zerolog"
)

type PermissionChecker struct {
	cfg        *config.Config
	httpClient *http.Client
	log        zerolog.Logger
}

func NewPermissionChecker(cfg *config.Config) *PermissionChecker {
	return &PermissionChecker{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.ExternalAPI.EducationDept.Timeout,
		},
		log: logger.Get(),
	}
}

// CheckPermission checks if score entry is allowed for the given class, semester, and year
func (p *PermissionChecker) CheckPermission(ctx context.Context, className string, semester string, year int) (bool, *model.SemesterPermissionResponse, error) {
	baseURL := p.cfg.ExternalAPI.EducationDept.BaseURL + p.cfg.ExternalAPI.EducationDept.SemesterPermissionEndpoint

	// Build query parameters
	params := url.Values{}
	params.Add("className", className)
	params.Add("semester", semester)
	params.Add("year", strconv.Itoa(year))

	fullURL := baseURL + "?" + params.Encode()

	p.log.Debug().
		Str("url", fullURL).
		Str("class", className).
		Str("semester", semester).
		Int("year", year).
		Msg("Checking semester permission")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check permission: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var permResp model.SemesterPermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&permResp); err != nil {
		return false, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log the result
	schedule := permResp.GetActiveSchedule()
	if schedule != nil {
		p.log.Info().
			Bool("is_allowed", permResp.IsAllowed).
			Str("class", className).
			Str("semester", semester).
			Int("year", year).
			Bool("active", schedule.Active).
			Bool("locked", schedule.Locked).
			Time("start", schedule.StartDateTime).
			Time("end", schedule.EndDateTime).
			Msg("Semester permission checked")
	} else {
		p.log.Info().
			Bool("is_allowed", permResp.IsAllowed).
			Str("class", className).
			Str("semester", semester).
			Int("year", year).
			Interface("active_schedule", permResp.ActiveSchedule).
			Msg("Semester permission checked - no active schedule")
	}

	return permResp.IsAllowed, &permResp, nil
}
