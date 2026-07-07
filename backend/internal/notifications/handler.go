package notifications

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	cfg *config.Config
	ns  *Service
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:  db,
		cfg: cfg,
		ns:  NewService(db, cfg),
	}
}

type CreateNotificationRequest struct {
	Title       string    `json:"title" binding:"required"`
	Message     string    `json:"message" binding:"required"`
	Type        string    `json:"type"`
	ScheduledAt string    `json:"scheduled_at"`
}

type ReminderRequest struct {
	Title       string `json:"title" binding:"required"`
	Message     string `json:"message" binding:"required"`
	ScheduledAt string `json:"scheduled_at" binding:"required"`
}

type PushNotificationRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type NotificationResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	Type        string     `json:"type"`
	IsRead      bool       `json:"is_read"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt   string     `json:"created_at"`
}

func (h *Handler) CreateNotification(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var scheduledAt *time.Time
	if req.ScheduledAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_at format"})
			return
		}
		scheduledAt = &parsedTime
	}

	notificationReq := NotificationRequest{
		UserID:      userID,
		Title:       req.Title,
		Message:     req.Message,
		Type:        req.Type,
		ScheduledAt: scheduledAt,
	}

	notification, err := h.ns.CreateNotification(notificationReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	response := NotificationResponse{
		ID:          notification.ID,
		Title:       notification.Title,
		Message:     notification.Message,
		Type:        notification.Type,
		IsRead:      notification.IsRead,
		ScheduledAt: notification.ScheduledAt,
		CreatedAt:   notification.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	notifications, err := h.ns.GetNotifications(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	var response []NotificationResponse
	for _, notif := range notifications {
		response = append(response, NotificationResponse{
			ID:          notif.ID,
			Title:       notif.Title,
			Message:     notif.Message,
			Type:        notif.Type,
			IsRead:      notif.IsRead,
			ScheduledAt: notif.ScheduledAt,
			CreatedAt:   notif.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetUnreadNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	notifications, err := h.ns.GetUnreadNotifications(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unread notifications"})
		return
	}

	var response []NotificationResponse
	for _, notif := range notifications {
		response = append(response, NotificationResponse{
			ID:          notif.ID,
			Title:       notif.Title,
			Message:     notif.Message,
			Type:        notif.Type,
			IsRead:      notif.IsRead,
			ScheduledAt: notif.ScheduledAt,
			CreatedAt:   notif.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	notificationID := c.Param("id")

	id, err := strconv.ParseUint(notificationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.ns.MarkAsRead(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := h.ns.MarkAllAsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

func (h *Handler) DeleteNotification(c *gin.Context) {
	userID := c.GetUint("user_id")
	notificationID := c.Param("id")

	id, err := strconv.ParseUint(notificationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.ns.DeleteNotification(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

func (h *Handler) CreateReminder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req ReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled_at format"})
		return
	}

	notification, err := h.ns.CreateReminder(userID, req.Title, req.Message, scheduledAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reminder"})
		return
	}

	response := NotificationResponse{
		ID:          notification.ID,
		Title:       notification.Title,
		Message:     notification.Message,
		Type:        notification.Type,
		IsRead:      notification.IsRead,
		ScheduledAt: notification.ScheduledAt,
		CreatedAt:   notification.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) SendPushNotification(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req PushNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ns.SendPushNotification(userID, req.Title, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send push notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Push notification sent successfully"})
}
