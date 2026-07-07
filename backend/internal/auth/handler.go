package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{db: db, cfg: cfg}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	User         User   `json:"user"`
}

type User struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is inactive"})
		return
	}

	token, err := GenerateToken(user.ID, user.Email, h.cfg.JWT)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := GenerateToken(user.ID, user.Email, h.cfg.JWT)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	response := LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(h.cfg.JWT.Expiration).Format(time.RFC3339),
		User: User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := ValidateToken(req.RefreshToken, h.cfg.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	token, err := GenerateToken(user.ID, user.Email, h.cfg.JWT)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := GenerateToken(user.ID, user.Email, config.JWTConfig{
		Secret:     h.cfg.JWT.Secret,
		Expiration: time.Hour * 24 * 30,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	response := LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(h.cfg.JWT.Expiration).Format(time.RFC3339),
		User: User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}

	c.JSON(http.StatusOK, response)
}
