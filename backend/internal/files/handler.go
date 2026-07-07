package files

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
	fs  *Service
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:  db,
		cfg: cfg,
		fs:  NewService(db, cfg),
	}
}

type QueryDocumentRequest struct {
	Query string `json:"query" binding:"required"`
}

type DocumentResponse struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	FileType  string `json:"file_type"`
	Status    string `json:"status"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) UploadDocument(c *gin.Context) {
	userID := c.GetUint("user_id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileReader.Close()

	document, err := h.fs.UploadDocument(userID, fileReader, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload document"})
		return
	}

	response := DocumentResponse{
		ID:        document.ID,
		Title:     document.Title,
		FileName:  document.FileName,
		FileSize:  document.FileSize,
		FileType:  document.FileType,
		Status:    document.Status,
		Summary:   document.Summary,
		CreatedAt: document.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetDocuments(c *gin.Context) {
	userID := c.GetUint("user_id")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	documents, err := h.fs.GetDocuments(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
		return
	}

	var response []DocumentResponse
	for _, doc := range documents {
		response = append(response, DocumentResponse{
			ID:        doc.ID,
			Title:     doc.Title,
			FileName:  doc.FileName,
			FileSize:  doc.FileSize,
			FileType:  doc.FileType,
			Status:    doc.Status,
			Summary:   doc.Summary,
			CreatedAt: doc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) AnalyzeDocument(c *gin.Context) {
	userID := c.GetUint("user_id")
	documentID := c.Param("id")

	id, err := strconv.ParseUint(documentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	analysis, err := h.fs.AnalyzeDocument(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze document"})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

func (h *Handler) QueryDocument(c *gin.Context) {
	userID := c.GetUint("user_id")
	documentID := c.Param("id")

	id, err := strconv.ParseUint(documentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var req QueryDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	answer, err := h.fs.QueryDocument(uint(id), userID, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	userID := c.GetUint("user_id")
	documentID := c.Param("id")

	id, err := strconv.ParseUint(documentID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	if err := h.fs.DeleteDocument(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}
