// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package code

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
)

type Handler struct {
	cfg *config.Config
	cs  *Service
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg: cfg,
		cs:  NewService(cfg),
	}
}

type ConvertCodeRequest struct {
	Code           string `json:"code" binding:"required"`
	SourceLanguage string `json:"source_language" binding:"required"`
	TargetLanguage string `json:"target_language" binding:"required"`
}

func (h *Handler) GenerateCode(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Language == "" {
		req.Language = h.cs.DetectLanguage(req.Query)
	}

	response, err := h.cs.GenerateCode(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) DebugCode(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.cs.DebugCode(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to debug code"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) ReviewCode(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := h.cs.ReviewCode(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to review code"})
		return
	}

	c.JSON(http.StatusOK, review)
}

func (h *Handler) OptimizeCode(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.cs.OptimizeCode(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to optimize code"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GenerateDocumentation(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.cs.GenerateDocumentation(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate documentation"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) ExplainCode(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.cs.ExplainCode(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to explain code"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) ConvertCode(c *gin.Context) {
	var req ConvertCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	codeReq := CodeRequest{
		Code:     req.Code,
		Language: req.SourceLanguage,
	}

	response, err := h.cs.ConvertCode(codeReq, req.TargetLanguage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert code"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GenerateTests(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.cs.GenerateTests(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tests"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetSupportedLanguages(c *gin.Context) {
	languages := h.cs.GetSupportedLanguages()
	c.JSON(http.StatusOK, gin.H{"languages": languages})
}
