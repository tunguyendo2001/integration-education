package api

import (
	"net/http"
	"strconv"

	"integration-education-db/internal/config"
	"integration-education-db/internal/db"
	"integration-education-db/internal/logger"
	"integration-education-db/internal/model"
	"integration-education-db/internal/queue"
	"integration-education-db/internal/sync"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Handler struct {
	repo        db.Repository
	producer    *queue.Producer
	syncService *sync.Service
	cfg         *config.Config
	log         zerolog.Logger
}

func NewHandler(
	repo db.Repository,
	producer *queue.Producer,
	cfg *config.Config,
) *Handler {
	return &Handler{
		repo:        repo,
		producer:    producer,
		syncService: sync.NewService(cfg, repo),
		cfg:         cfg,
		log:         logger.Get(),
	}
}

func (h *Handler) TriggerSync(c *gin.Context) {
	var req model.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate file exists and is parsed
	file, err := h.repo.GetFile(c.Request.Context(), req.FileID)
	if err != nil {
		h.log.Error().Err(err).Int64("file_id", req.FileID).Msg("File not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if file.Status != model.FileStatusParsedOK {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "File is not ready for sync",
			"status": file.Status,
		})
		return
	}

	// Check sync window
	withinWindow, err := h.syncService.IsWithinSyncWindow()
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to check sync window")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if !withinWindow {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Sync not allowed at this time",
			"sync_window": gin.H{
				"start": h.cfg.SyncWindow.StartTime,
				"end":   h.cfg.SyncWindow.EndTime,
				"tz":    h.cfg.SyncWindow.Timezone,
			},
		})
		return
	}

	// Enqueue sync job
	job := model.SyncJob{
		FileID:  req.FileID,
		Class:   req.Class,
		Subject: req.Subject,
	}

	if err := h.producer.EnqueueSyncJob(c.Request.Context(), job); err != nil {
		h.log.Error().Err(err).Msg("Failed to enqueue sync job")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue sync job"})
		return
	}

	h.log.Info().
		Int64("file_id", req.FileID).
		Str("class", req.Class).
		Str("subject", req.Subject).
		Msg("Sync job enqueued")

	c.JSON(http.StatusOK, gin.H{
		"message": "Sync job queued successfully",
		"job":     job,
	})
}

func (h *Handler) GetGradesStatus(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	status, err := h.repo.GetGradesStatus(c.Request.Context(), fileID)
	if err != nil {
		h.log.Error().Err(err).Int64("file_id", fileID).Msg("Failed to get grades status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get file info
	file, err := h.repo.GetFile(c.Request.Context(), fileID)
	if err != nil {
		h.log.Error().Err(err).Int64("file_id", fileID).Msg("File not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	status.Status = string(file.Status)
	c.JSON(http.StatusOK, status)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": h.cfg.App.Name,
		"version": h.cfg.App.Version,
	})
}
