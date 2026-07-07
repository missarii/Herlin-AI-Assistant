// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package voice

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

type TranscriptionRequest struct {
	AudioData   []byte `json:"audio_data" binding:"required"`
	AudioFormat string `json:"audio_format" binding:"required"`
}

type SynthesisRequest struct {
	Text     string `json:"text" binding:"required"`
	Voice    string `json:"voice"`
	Language string `json:"language"`
}

func (h *Handler) SpeechToText(c *gin.Context) {
	var req TranscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.vs.SpeechToText(req.AudioData, req.AudioFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transcribe audio"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) TextToSpeech(c *gin.Context) {
	var req SynthesisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.vs.TextToSpeech(req.Text, req.Voice, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to synthesize speech"})
		return
	}

	// Return audio data directly
	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Disposition", "attachment; filename=speech.mp3")
	c.Data(http.StatusOK, "audio/mpeg", result.AudioData)
}

func (h *Handler) ProcessVoiceCommand(c *gin.Context) {
	var req TranscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// First transcribe
	transcription, err := h.vs.SpeechToText(req.AudioData, req.AudioFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transcribe audio"})
		return
	}

	// Process command
	command, text, err := h.vs.ProcessVoiceCommand(transcription.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process command"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"command":      command,
		"transcription": transcription.Text,
		"text":         text,
	})
}
