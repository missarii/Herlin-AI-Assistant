package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	cfg *config.Config
}

type MemoryRequest struct {
	UserID    uint   `json:"user_id"`
	Content   string `json:"content"`
	Importance float32 `json:"importance"`
}

type SearchRequest struct {
	UserID  uint   `json:"user_id"`
	Query   string `json:"query"`
	Limit   int    `json:"limit"`
}

type SearchResult struct {
	Content     string  `json:"content"`
	Importance  float32 `json:"importance"`
	Similarity  float32 `json:"similarity"`
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

func (s *Service) StoreMemory(userID uint, content string, importance float32) error {
	// Generate embedding
	embedding, err := s.generateEmbedding(content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Store in database
	memory := models.Memory{
		UserID:     userID,
		Content:    content,
		Embedding:  embedding,
		Importance: importance,
		AccessCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.db.Create(&memory).Error; err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}

	// Store in Qdrant
	if err := s.storeInQdrant(memory); err != nil {
		// Log error but don't fail - memory is still in PostgreSQL
		fmt.Printf("Warning: failed to store in Qdrant: %v\n", err)
	}

	return nil
}

func (s *Service) SearchMemories(userID uint, query string, limit int) ([]SearchResult, error) {
	// Generate query embedding
	queryEmbedding, err := s.generateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search in Qdrant
	qdrantResults, err := s.searchInQdrant(userID, queryEmbedding, limit)
	if err != nil {
		// Fallback to PostgreSQL full-text search
		return s.searchInPostgreSQL(userID, query, limit)
	}

	return qdrantResults, nil
}

func (s *Service) GetMemories(userID uint, limit int) ([]models.Memory, error) {
	var memories []models.Memory
	err := s.db.Where("user_id = ?", userID).
		Order("importance DESC, access_count DESC").
		Limit(limit).
		Find(&memories).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch memories: %w", err)
	}

	return memories, nil
}

func (s *Service) DeleteMemory(memoryID uint, userID uint) error {
	var memory models.Memory
	if err := s.db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
		return fmt.Errorf("memory not found: %w", err)
	}

	if err := s.db.Delete(&memory).Error; err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	// Delete from Qdrant
	if err := s.deleteFromQdrant(memoryID); err != nil {
		fmt.Printf("Warning: failed to delete from Qdrant: %v\n", err)
	}

	return nil
}

func (s *Service) UpdateMemoryImportance(memoryID uint, userID uint, importance float32) error {
	var memory models.Memory
	if err := s.db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
		return fmt.Errorf("memory not found: %w", err)
	}

	memory.Importance = importance
	memory.UpdatedAt = time.Now()

	if err := s.db.Save(&memory).Error; err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	return nil
}

// Embedding generation using OpenAI
func (s *Service) generateEmbedding(text string) ([]float32, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		// Return dummy embedding if no API key
		return make([]float32, 1536), nil
	}

	type EmbeddingRequest struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}

	type EmbeddingResponse struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	req := EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: text,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.OpenAI.APIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embeddingResp.Data[0].Embedding, nil
}

// Qdrant operations
type QdrantPoint struct {
	ID      uint      `json:"id"`
	Vector  []float32 `json:"vector"`
	Payload struct {
		UserID    uint   `json:"user_id"`
		Content   string `json:"content"`
		Importance float32 `json:"importance"`
	} `json:"payload"`
}

type QdrantUpsertRequest struct {
	Points []QdrantPoint `json:"points"`
}

func (s *Service) storeInQdrant(memory models.Memory) error {
	if memory.Embedding == nil {
		return fmt.Errorf("no embedding available")
	}

	point := QdrantPoint{
		ID:     memory.ID,
		Vector: memory.Embedding,
	}
	point.Payload.UserID = memory.UserID
	point.Payload.Content = memory.Content
	point.Payload.Importance = memory.Importance

	req := QdrantUpsertRequest{
		Points: []QdrantPoint{point},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s:%s/collections/memories/points", s.cfg.Qdrant.Host, s.cfg.Qdrant.Port)
	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Qdrant API error: %s", string(body))
	}

	return nil
}

type QdrantSearchRequest struct {
	Vector  []float32 `json:"vector"`
	Limit   int       `json:"limit"`
	Filter  struct {
		Must []struct {
			Key   string `json:"key"`
			Match struct {
				Value uint `json:"value"`
			} `json:"match"`
		} `json:"must"`
	} `json:"filter"`
}

type QdrantSearchResponse struct {
	Result []struct {
		ID      uint `json:"id"`
		Score   float32 `json:"score"`
		Payload struct {
			Content    string  `json:"content"`
			Importance float32 `json:"importance"`
		} `json:"payload"`
	} `json:"result"`
}

func (s *Service) searchInQdrant(userID uint, embedding []float32, limit int) ([]SearchResult, error) {
	req := QdrantSearchRequest{
		Vector: embedding,
		Limit:  limit,
	}
	req.Filter.Must = append(req.Filter.Must, struct {
		Key   string `json:"key"`
		Match struct {
			Value uint `json:"value"`
		} `json:"match"`
	}{
		Key: "user_id",
		Match: struct {
			Value uint `json:"value"`
		}{Value: userID},
	})

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s:%s/collections/memories/points/search", s.cfg.Qdrant.Host, s.cfg.Qdrant.Port)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Qdrant API error: %s", string(body))
	}

	var qdrantResp QdrantSearchResponse
	if err := json.Unmarshal(body, &qdrantResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var results []SearchResult
	for _, result := range qdrantResp.Result {
		results = append(results, SearchResult{
			Content:    result.Payload.Content,
			Importance: result.Payload.Importance,
			Similarity: result.Score,
		})

		// Update access count
		s.db.Model(&models.Memory{}).Where("id = ?", result.ID).UpdateColumn("access_count", gorm.Expr("access_count + ?", 1))
	}

	return results, nil
}

func (s *Service) searchInPostgreSQL(userID uint, query string, limit int) ([]SearchResult, error) {
	var memories []models.Memory
	err := s.db.Where("user_id = ? AND content ILIKE ?", userID, "%"+query+"%").
		Order("importance DESC").
		Limit(limit).
		Find(&memories).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	var results []SearchResult
	for _, memory := range memories {
		results = append(results, SearchResult{
			Content:    memory.Content,
			Importance: memory.Importance,
			Similarity: 0.5, // Default similarity for PostgreSQL search
		})

		// Update access count
		s.db.Model(&memory).UpdateColumn("access_count", gorm.Expr("access_count + ?", 1))
	}

	return results, nil
}

func (s *Service) deleteFromQdrant(memoryID uint) error {
	url := fmt.Sprintf("http://%s:%s/collections/memories/points/delete", s.cfg.Qdrant.Host, s.cfg.Qdrant.Port)
	
	req := struct {
		Points []uint `json:"points"`
	}{
		Points: []uint{memoryID},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Qdrant API error")
	}

	return nil
}

func (s *Service) InitializeQdrantCollection() error {
	type CreateCollectionRequest struct {
		Vector struct {
			Size     uint   `json:"size"`
			Distance string `json:"distance"`
		} `json:"vector"`
	}

	req := CreateCollectionRequest{}
	req.Vector.Size = 1536 // OpenAI embedding size
	req.Vector.Distance = "cosine"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s:%s/collections/memories", s.cfg.Qdrant.Host, s.cfg.Qdrant.Port)
	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Ignore errors if collection already exists
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Warning: failed to create Qdrant collection: %s\n", string(body))
	}

	return nil
}
