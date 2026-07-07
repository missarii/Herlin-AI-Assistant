// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package vision

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
)

type Handler struct {
	cfg *config.Config
	vs  *Service
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
		vs:  NewService(cfg),
	}
}

type AnalyzeImageRequest struct {
	ImageData   []byte `json:"image_data" binding:"required"`
	ImageFormat string `json:"image_format" binding:"required"`
	Query       string `json:"query"`
}

func (h *Handler) AnalyzeImage(c *gin.Context) {
	var req AnalyzeImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analysis, err := h.vs.AnalyzeImage(req.ImageData, req.ImageFormat, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze image"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

func (h *Handler) AnalyzeScreenshot(c *gin.Context) {
	var req AnalyzeImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analysis, err := h.vs.AnalyzeScreenshot(req.ImageData, req.ImageFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze screenshot"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

func (h *Handler) AnalyzeDiagram(c *gin.Context) {
	var req AnalyzeImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analysis, err := h.vs.AnalyzeDiagram(req.ImageData, req.ImageFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze diagram"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}
