package chat

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/ai"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	cfg *config.Config
	ai  *ai.Service
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:  db,
		cfg: cfg,
		ai:  ai.NewService(cfg),
	}
}

type SendMessageRequest struct {
	ConversationID *uint   `json:"conversation_id,omitempty"`
	Message        string `json:"message" binding:"required"`
	Model          string `json:"model,omitempty"`
}

type SendMessageResponse struct {
	MessageID       uint   `json:"message_id"`
	ConversationID  uint   `json:"conversation_id"`
	Role            string `json:"role"`
	Content         string `json:"content"`
	Tokens          int    `json:"tokens"`
	CreatedAt       string `json:"created_at"`
}

type CreateConversationRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
}

type ConversationResponse struct {
	ID         uint                  `json:"id"`
	Title      string                `json:"title"`
	Model      string                `json:"model"`
	IsArchived bool                  `json:"is_archived"`
	CreatedAt  string                `json:"created_at"`
	UpdatedAt  string                `json:"updated_at"`
}

type MessageResponse struct {
	ID        uint   `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Tokens    int    `json:"tokens"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) SendMessage(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var conversation models.Conversation
	var conversationID uint

	if req.ConversationID != nil {
		// Use existing conversation
		if err := h.db.Where("id = ? AND user_id = ?", *req.ConversationID, userID).First(&conversation).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
			return
		}
		conversationID = conversation.ID
	} else {
		// Create new conversation
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

	// Get conversation history for context
	var messages []models.Message
	h.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages)

	// Convert to AI format
	var aiMessages []ai.Message
	for _, msg := range messages {
		aiMessages = append(aiMessages, ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Get AI response
	response, err := h.ai.GenerateResponse(aiMessages, conversation.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate AI response"})
		return
	}

	// Save assistant message
	assistantMessage := models.Message{
		ConversationID: conversationID,
		Role:          "assistant",
		Content:       response.Content,
		Tokens:        response.Tokens,
		CreatedAt:     time.Now(),
	}
	if err := h.db.Create(&assistantMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save assistant message"})
		return
	}

	// Update conversation title if it's the first message
	if len(messages) == 1 {
		conversation.Title = generateTitle(req.Message)
		h.db.Save(&conversation)
	}

	resp := SendMessageResponse{
		MessageID:      assistantMessage.ID,
		ConversationID: conversationID,
		Role:          assistantMessage.Role,
		Content:       assistantMessage.Content,
		Tokens:        assistantMessage.Tokens,
		CreatedAt:     assistantMessage.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetConversations(c *gin.Context) {
	userID := c.GetUint("user_id")

	var conversations []models.Conversation
	if err := h.db.Where("user_id = ? AND deleted_at IS NULL", userID).Order("updated_at DESC").Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	var response []ConversationResponse
	for _, conv := range conversations {
		response = append(response, ConversationResponse{
			ID:         conv.ID,
			Title:      conv.Title,
			Model:      conv.Model,
			IsArchived: conv.IsArchived,
			CreatedAt:  conv.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  conv.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateConversation(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation := models.Conversation{
		UserID: userID,
		Title:  req.Title,
		Model:  h.getModel(req.Model),
	}

	if err := h.db.Create(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	response := ConversationResponse{
		ID:         conversation.ID,
		Title:      conversation.Title,
		Model:      conversation.Model,
		IsArchived: conversation.IsArchived,
		CreatedAt:  conversation.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  conversation.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetConversation(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var conversation models.Conversation
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	response := ConversationResponse{
		ID:         conversation.ID,
		Title:      conversation.Title,
		Model:      conversation.Model,
		IsArchived: conversation.IsArchived,
		CreatedAt:  conversation.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  conversation.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) UpdateConversation(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var conversation models.Conversation
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	if req.Title != "" {
		conversation.Title = req.Title
	}
	if req.Model != "" {
		conversation.Model = req.Model
	}

	if err := h.db.Save(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conversation"})
		return
	}

	response := ConversationResponse{
		ID:         conversation.ID,
		Title:      conversation.Title,
		Model:      conversation.Model,
		IsArchived: conversation.IsArchived,
		CreatedAt:  conversation.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  conversation.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteConversation(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var conversation models.Conversation
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	if err := h.db.Delete(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted successfully"})
}

func (h *Handler) GetMessages(c *gin.Context) {
	userID := c.GetUint("user_id")
	conversationID := c.Param("id")

	// Verify conversation belongs to user
	var conversation models.Conversation
	if err := h.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	var messages []models.Message
	if err := h.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	var response []MessageResponse
	for _, msg := range messages {
		response = append(response, MessageResponse{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			Tokens:    msg.Tokens,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getModel(model string) string {
	if model == "" {
		return h.cfg.AI.DefaultProvider
	}
	return model
}

func generateTitle(message string) string {
	// Simple title generation - take first 50 characters
	if len(message) > 50 {
		return message[:50] + "..."
	}
	return message
}
