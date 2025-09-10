package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"integration-education-db/internal/config"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"

	"github.com/rs/zerolog"
)

type AuthManager struct {
	cfg       *config.Config
	client    *http.Client
	token     string
	expiresAt time.Time
	mu        sync.RWMutex
	log       zerolog.Logger
}

func NewAuthManager(cfg *config.Config) *AuthManager {
	return &AuthManager{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: logger.Get(),
	}
}

func (a *AuthManager) GetToken(ctx context.Context) (string, error) {
	a.mu.RLock()
	if a.token != "" && time.Now().Before(a.expiresAt.Add(-30*time.Second)) {
		token := a.token
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	return a.refreshToken(ctx)
}

func (a *AuthManager) refreshToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double check after acquiring write lock
	if a.token != "" && time.Now().Before(a.expiresAt.Add(-30*time.Second)) {
		return a.token, nil
	}

	a.log.Debug().Msg("Refreshing authentication token")

	authData := map[string]string{
		"username": a.cfg.ExternalAPI.EducationDept.Username,
		"password": a.cfg.ExternalAPI.EducationDept.Password,
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth data: %w", err)
	}

	url := a.cfg.ExternalAPI.EducationDept.BaseURL + a.cfg.ExternalAPI.EducationDept.AuthEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth failed with status: %d", resp.StatusCode)
	}

	var tokenResp model.AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	a.token = tokenResp.Token
	a.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	a.log.Debug().Time("expires_at", a.expiresAt).Msg("Token refreshed successfully")

	return a.token, nil
}
