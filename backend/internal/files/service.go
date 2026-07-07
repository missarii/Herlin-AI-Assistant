package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
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

type DocumentAnalysis struct {
	Summary     string   `json:"summary"`
	KeyPoints   []string `json:"key_points"`
	Topics      []string `json:"topics"`
	Language    string   `json:"language"`
	WordCount   int      `json:"word_count"`
	PageCount   int      `json:"page_count,omitempty"`
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg, ai: ai.NewService(cfg)}
}

func (s *Service) UploadDocument(userID uint, file multipart.File, header *multipart.FileHeader) (*models.Document, error) {
	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Store file
	filePath, err := s.storeFile(userID, header.Filename, content)
	if err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	// Create document record
	document := models.Document{
		UserID:    userID,
		Title:     header.Filename,
		FileName:  header.Filename,
		FilePath:  filePath,
		FileSize:  header.Size,
		FileType:  getFileType(header.Filename),
		Status:    "processing",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(&document).Error; err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// Process document asynchronously
	go s.processDocument(&document, content)

	return &document, nil
}

func (s *Service) processDocument(document *models.Document, content []byte) {
	// Extract text from document based on file type
	var text string
	var err error

	switch document.FileType {
	case "pdf":
		text, err = s.extractTextFromPDF(content)
	case "docx":
		text, err = s.extractTextFromDocx(content)
	case "txt":
		text = string(content)
	default:
		text = string(content)
	}

	if err != nil {
		document.Status = "failed"
		s.db.Save(document)
		return
	}

	// Generate summary using AI
	summary, err := s.generateDocumentSummary(text)
	if err != nil {
		document.Status = "failed"
		s.db.Save(document)
		return
	}

	// Update document
	document.Summary = summary
	document.Status = "completed"
	document.UpdatedAt = time.Now()

	if err := s.db.Save(document).Error; err != nil {
		fmt.Printf("Failed to update document: %v\n", err)
	}
}

func (s *Service) generateDocumentSummary(text string) (string, error) {
	// Truncate text if too long
	maxLength := 10000
	if len(text) > maxLength {
		text = text[:maxLength]
	}

	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a document analysis assistant. Provide a concise summary of the given document.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Please summarize this document:\n\n%s", text),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

func (s *Service) AnalyzeDocument(documentID uint, userID uint) (*DocumentAnalysis, error) {
	var document models.Document
	if err := s.db.Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Read file content
	content, err := s.readFile(document.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Extract text
	var text string
	switch document.FileType {
	case "pdf":
		text, err = s.extractTextFromPDF(content)
	case "docx":
		text, err = s.extractTextFromDocx(content)
	default:
		text = string(content)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Generate analysis
	analysis, err := s.generateAnalysis(text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}

	return analysis, nil
}

func (s *Service) generateAnalysis(text string) (*DocumentAnalysis, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a document analysis assistant. Analyze the document and provide a structured JSON response with summary, key points, topics, language, and word count.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Analyze this document and provide a JSON response:\n\n%s", text),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var analysis DocumentAnalysis
	if err := json.Unmarshal([]byte(response.Content), &analysis); err != nil {
		// Fallback if AI doesn't return valid JSON
		analysis = DocumentAnalysis{
			Summary:   response.Content,
			WordCount: len(strings.Fields(text)),
		}
	}

	return &analysis, nil
}

func (s *Service) QueryDocument(documentID uint, userID uint, query string) (string, error) {
	var document models.Document
	if err := s.db.Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		return "", fmt.Errorf("document not found: %w", err)
	}

	// Read file content
	content, err := s.readFile(document.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Extract text
	var text string
	switch document.FileType {
	case "pdf":
		text, err = s.extractTextFromPDF(content)
	case "docx":
		text, err = s.extractTextFromDocx(content)
	default:
		text = string(content)
	}

	if err != nil {
		return "", fmt.Errorf("failed to extract text: %w", err)
	}

	// Query AI with document context
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a document Q&A assistant. Answer questions based on the provided document content.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Document content:\n\n%s\n\nQuestion: %s", text, query),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

func (s *Service) storeFile(userID uint, filename string, content []byte) (string, error) {
	// For now, store locally. In production, use S3/MinIO
	extension := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, extension)
	newFilename := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), extension)
	filePath := filepath.Join("uploads", newFilename)

	// Create uploads directory if it doesn't exist
	// os.MkdirAll("uploads", 0755)

	// Write file
	// if err := os.WriteFile(filePath, content, 0644); err != nil {
	// 	return "", err
	// }

	// For now, return a mock path
	return filePath, nil
}

func (s *Service) readFile(filePath string) ([]byte, error) {
	// Read file from storage
	// return os.ReadFile(filePath)
	
	// For now, return empty
	return []byte{}, nil
}

func (s *Service) extractTextFromPDF(content []byte) (string, error) {
	// Use a PDF parsing library like pdf-to-text or unidoc
	// For now, return placeholder
	return "PDF text extraction not implemented", nil
}

func (s *Service) extractTextFromDocx(content []byte) (string, error) {
	// Use a DOCX parsing library
	// For now, return placeholder
	return "DOCX text extraction not implemented", nil
}

func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "docx"
	case ".txt":
		return "txt"
	case ".md":
		return "markdown"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	default:
		return "unknown"
	}
}

func (s *Service) GetDocuments(userID uint, limit int) ([]models.Document, error) {
	var documents []models.Document
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&documents).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch documents: %w", err)
	}

	return documents, nil
}

func (s *Service) DeleteDocument(documentID uint, userID uint) error {
	var document models.Document
	if err := s.db.Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	if err := s.db.Delete(&document).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Delete file from storage
	// os.Remove(document.FilePath)

	return nil
}
