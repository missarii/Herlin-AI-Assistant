package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/internal/ai"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"gorm.io/gorm"
)

type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Token   int    `json:"token,omitempty"`
}

func (h *Handler) SendMessageStream(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var conversation models.Conversation
	var conversationID uint

	if req.ConversationID != nil {
		if err := h.db.Where("id = ? AND user_id = ?", *req.ConversationID, userID).First(&conversation).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
			return
		}
		conversationID = conversation.ID
	} else {
		conversation = models.Conversation{
			UserID: userID,
			Title:  "New Chat",
			Model:  h.getModel(req.Model),
		}
		if err := h.db.Create(&conversation).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
			return
		}
		conversationID = conversation.ID
	}

	// Save user message
	userMessage := models.Message{
		ConversationID: conversationID,
		Role:          "user",
		Content:       req.Message,
		CreatedAt:     time.Now(),
	}
	if err := h.db.Create(&userMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	// Get conversation history
	var messages []models.Message
	h.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages)

	var aiMessages []ai.Message
	for _, msg := range messages {
		aiMessages = append(aiMessages, ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Create channel for streaming
	streamChan := make(chan StreamChunk)
	errChan := make(chan error)

	// Start streaming in goroutine
	go func() {
		defer close(streamChan)
		defer close(errChan)

		response, err := h.ai.GenerateResponseStream(aiMessages, conversation.Model, streamChan)
		if err != nil {
			errChan <- err
			return
		}

		// Save complete assistant message
		assistantMessage := models.Message{
			ConversationID: conversationID,
			Role:          "assistant",
			Content:       response.Content,
			Tokens:        response.Tokens,
			CreatedAt:     time.Now(),
		}
		if err := h.db.Create(&assistantMessage).Error; err != nil {
			errChan <- err
			return
		}

		// Update conversation title if first message
		if len(messages) == 1 {
			conversation.Title = generateTitle(req.Message)
			h.db.Save(&conversation)
		}
	}()

	// Stream to client
	c.Stream(func(w io.Writer) bool {
		select {
		case chunk := <-streamChan:
			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			return !chunk.Done
		case err := <-errChan:
			if err != nil {
				errorChunk := StreamChunk{
					Content: fmt.Sprintf("Error: %s", err.Error()),
					Done:    true,
				}
				data, _ := json.Marshal(errorChunk)
				fmt.Fprintf(w, "data: %s\n\n", data)
			}
			return false
		case <-c.Request.Context().Done():
			return false
		}
	})
}
