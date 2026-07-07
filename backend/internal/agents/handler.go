package agents

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	cfg *config.Config
	as  *Service
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:  db,
		cfg: cfg,
		as:  NewService(db, cfg),
	}
}

type CreateAgentRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Capabilities []string             `json:"capabilities"`
	Settings    map[string]interface{} `json:"settings"`
}

type ExecuteAgentRequest struct {
	Task string `json:"task" binding:"required"`
}

type AgentResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      string `json:"config"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (h *Handler) CreateAgent(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := AgentConfig{
		Name:        req.Name,
		Description: req.Description,
		Capabilities: req.Capabilities,
		Settings:    req.Settings,
	}

	agent, err := h.as.CreateAgent(userID, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	response := AgentResponse{
		ID:          agent.ID,
		Name:        agent.Name,
		Description: agent.Description,
		Config:      agent.Config,
		IsActive:    agent.IsActive,
		CreatedAt:   agent.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   agent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetAgents(c *gin.Context) {
	userID := c.GetUint("user_id")

	agents, err := h.as.GetAgents(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	var response []AgentResponse
	for _, agent := range agents {
		response = append(response, AgentResponse{
			ID:          agent.ID,
			Name:        agent.Name,
			Description: agent.Description,
			Config:      agent.Config,
			IsActive:    agent.IsActive,
			CreatedAt:   agent.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   agent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetAgent(c *gin.Context) {
	userID := c.GetUint("user_id")
	agentID := c.Param("id")

	id, err := strconv.ParseUint(agentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	agent, err := h.as.GetAgent(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	response := AgentResponse{
		ID:          agent.ID,
		Name:        agent.Name,
		Description: agent.Description,
		Config:      agent.Config,
		IsActive:    agent.IsActive,
		CreatedAt:   agent.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   agent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) UpdateAgent(c *gin.Context) {
	userID := c.GetUint("user_id")
	agentID := c.Param("id")

	id, err := strconv.ParseUint(agentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := AgentConfig{
		Name:        req.Name,
		Description: req.Description,
		Capabilities: req.Capabilities,
		Settings:    req.Settings,
	}

	if err := h.as.UpdateAgent(uint(id), userID, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent updated successfully"})
}

func (h *Handler) DeleteAgent(c *gin.Context) {
	userID := c.GetUint("user_id")
	agentID := c.Param("id")

	id, err := strconv.ParseUint(agentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	if err := h.as.DeleteAgent(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}

func (h *Handler) ExecuteAgent(c *gin.Context) {
	userID := c.GetUint("user_id")
	agentID := c.Param("id")

	id, err := strconv.ParseUint(agentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var req ExecuteAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.as.ExecuteAgent(uint(id), userID, req.Task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute agent"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")

	task, err := h.as.GetTaskStatus(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}
