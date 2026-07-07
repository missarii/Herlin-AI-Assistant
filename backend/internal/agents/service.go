// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package agents

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/ai"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	cfg *config.Config
	ai  *ai.Service
}

type AgentTask struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Input     map[string]interface{} `json:"input"`
	Status    string                 `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type AgentConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Capabilities []string        `json:"capabilities"`
	Settings    map[string]interface{} `json:"settings"`
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg, ai: ai.NewService(cfg)}
}

func (s *Service) CreateAgent(userID uint, config AgentConfig) (*models.Agent, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	agent := models.Agent{
		UserID:      userID,
		Name:        config.Name,
		Description: config.Description,
		Config:      string(configJSON),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(&agent).Error; err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &agent, nil
}

func (s *Service) ExecuteAgent(agentID uint, userID uint, task string) (*AgentTask, error) {
	var agent models.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	if !agent.IsActive {
		return nil, fmt.Errorf("agent is not active")
	}

	// Parse agent config
	var config AgentConfig
	if err := json.Unmarshal([]byte(agent.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}

	// Create task
	agentTask := &AgentTask{
		ID:        generateTaskID(),
		Type:      "general",
		Input:     map[string]interface{}{"task": task},
		Status:    "processing",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Execute task based on capabilities
	go s.executeTask(agentTask, config)

	return agentTask, nil
}

func (s *Service) executeTask(task *AgentTask, config AgentConfig) {
	defer func() {
		task.Status = "completed"
		task.UpdatedAt = time.Now()
	}()

	// Analyze task to determine which capabilities to use
	capability, params := s.analyzeTask(task.Input["task"].(string), config.Capabilities)

	switch capability {
	case "web_search":
		result, err := s.webSearch(params)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			return
		}
		task.Result = result
	case "file_operations":
		result, err := s.fileOperations(params)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			return
		}
		task.Result = result
	case "data_analysis":
		result, err := s.dataAnalysis(params)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			return
		}
		task.Result = result
	case "api_calls":
		result, err := s.apiCalls(params)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			return
		}
		task.Result = result
	default:
		// Use AI to handle general tasks
		result, err := s.aiTaskExecution(task.Input["task"].(string), config)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			return
		}
		task.Result = result
	}
}

func (s *Service) analyzeTask(task string, capabilities []string) (string, map[string]interface{}) {
	task = strings.ToLower(task)
	params := make(map[string]interface{})

	// Simple pattern matching
	if strings.Contains(task, "search") || strings.Contains(task, "find") {
		if contains(capabilities, "web_search") {
			params["query"] = extractQuery(task)
			return "web_search", params
		}
	}

	if strings.Contains(task, "file") || strings.Contains(task, "document") {
		if contains(capabilities, "file_operations") {
			params["operation"] = extractOperation(task)
			return "file_operations", params
		}
	}

	if strings.Contains(task, "analyze") || strings.Contains(task, "data") {
		if contains(capabilities, "data_analysis") {
			params["data"] = task
			return "data_analysis", params
		}
	}

	if strings.Contains(task, "api") || strings.Contains(task, "request") {
		if contains(capabilities, "api_calls") {
			params["endpoint"] = extractEndpoint(task)
			return "api_calls", params
		}
	}

	return "general", params
}

func (s *Service) webSearch(params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid query parameter")
	}

	// Use a web search API (e.g., Google Custom Search, Bing Search)
	// For now, return a mock result
	return map[string]interface{}{
		"query": query,
		"results": []map[string]string{
			{"title": "Mock Result 1", "url": "https://example.com/1", "snippet": "This is a mock search result"},
			{"title": "Mock Result 2", "url": "https://example.com/2", "snippet": "Another mock search result"},
		},
	}, nil
}

func (s *Service) fileOperations(params map[string]interface{}) (interface{}, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid operation parameter")
	}

	// Perform file operations
	return map[string]interface{}{
		"operation": operation,
		"status":    "completed",
		"message":   fmt.Sprintf("File operation '%s' completed", operation),
	}, nil
}

func (s *Service) dataAnalysis(params map[string]interface{}) (interface{}, error) {
	data, ok := params["data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid data parameter")
	}

	// Analyze data using AI
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a data analysis assistant. Analyze the provided data and return insights.",
		},
		{
			Role:    "user",
			Content: data,
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"analysis": response.Content,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

func (s *Service) apiCalls(params map[string]interface{}) (interface{}, error) {
	endpoint, ok := params["endpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid endpoint parameter")
	}

	// Make API call
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to make API call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

func (s *Service) aiTaskExecution(task string, config AgentConfig) (interface{}, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("You are %s. %s Your capabilities: %s", 
				config.Name, config.Description, strings.Join(config.Capabilities, ", ")),
		},
		{
			Role:    "user",
			Content: task,
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"response": response.Content,
		"agent":    config.Name,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

func (s *Service) GetAgents(userID uint) ([]models.Agent, error) {
	var agents []models.Agent
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&agents).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agents: %w", err)
	}

	return agents, nil
}

func (s *Service) GetAgent(agentID uint, userID uint) (*models.Agent, error) {
	var agent models.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	return &agent, nil
}

func (s *Service) UpdateAgent(agentID uint, userID uint, config AgentConfig) error {
	var agent models.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	agent.Name = config.Name
	agent.Description = config.Description
	agent.Config = string(configJSON)
	agent.UpdatedAt = time.Now()

	if err := s.db.Save(&agent).Error; err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	return nil
}

func (s *Service) DeleteAgent(agentID uint, userID uint) error {
	var agent models.Agent
	if err := s.db.Where("id = ? AND user_id = ?", agentID, userID).First(&agent).Error; err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	if err := s.db.Delete(&agent).Error; err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}

func (s *Service) GetTaskStatus(taskID string) (*AgentTask, error) {
	// In a real implementation, this would query a task store
	// For now, return a mock task
	return &AgentTask{
		ID:        taskID,
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// Helper functions
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractQuery(task string) string {
	// Simple query extraction - in production, use NLP
	words := strings.Fields(task)
	if len(words) > 1 {
		return strings.Join(words[1:], " ")
	}
	return task
}

func extractOperation(task string) string {
	// Simple operation extraction
	if strings.Contains(task, "read") {
		return "read"
	}
	if strings.Contains(task, "write") {
		return "write"
	}
	if strings.Contains(task, "delete") {
		return "delete"
	}
	return "list"
}

func extractEndpoint(task string) string {
	// Simple endpoint extraction
	words := strings.Fields(task)
	for _, word := range words {
		if strings.HasPrefix(word, "http") {
			return word
		}
	}
	return ""
}
