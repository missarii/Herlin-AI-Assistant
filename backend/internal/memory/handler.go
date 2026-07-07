package memory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	cfg *config.Config
	mem *Service
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:  db,
		cfg: cfg,
		mem: NewService(db, cfg),
	}
}

type StoreMemoryRequest struct {
	Content   string  `json:"content" binding:"required"`
	Importance float32 `json:"importance" binding:"required,min=0,max=1"`
}

type SearchMemoryRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit"`
}

type MemoryResponse struct {
	ID          uint      `json:"id"`
	Content     string    `json:"content"`
	Importance  float32   `json:"importance"`
	AccessCount int       `json:"access_count"`
	CreatedAt   string    `json:"created_at"`
}

func (h *Handler) StoreMemory(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req StoreMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.mem.StoreMemory(userID, req.Content, req.Importance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store memory"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Memory stored successfully"})
}

func (h *Handler) SearchMemories(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req SearchMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	results, err := h.mem.SearchMemories(userID, req.Query, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search memories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (h *Handler) GetMemories(c *gin.Context) {
	userID := c.GetUint("user_id")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	memories, err := h.mem.GetMemories(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch memories"})
		return
	}

	var response []MemoryResponse
	for _, memory := range memories {
		response = append(response, MemoryResponse{
			ID:          memory.ID,
			Content:     memory.Content,
			Importance:  memory.Importance,
			AccessCount: memory.AccessCount,
			CreatedAt:   memory.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteMemory(c *gin.Context) {
	userID := c.GetUint("user_id")
	memoryID := c.Param("id")

	id, err := strconv.ParseUint(memoryID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	if err := h.mem.DeleteMemory(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete memory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memory deleted successfully"})
}

func (h *Handler) UpdateMemoryImportance(c *gin.Context) {
	userID := c.GetUint("user_id")
	memoryID := c.Param("id")

	id, err := strconv.ParseUint(memoryID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid memory ID"})
		return
	}

	var req StoreMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.mem.UpdateMemoryImportance(uint(id), userID, req.Importance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update memory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memory updated successfully"})
}
