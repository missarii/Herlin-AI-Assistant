package notifications

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	cfg *config.Config
}

type NotificationRequest struct {
	UserID      uint      `json:"user_id"`
	Title       string    `json:"title" binding:"required"`
	Message     string    `json:"message" binding:"required"`
	Type        string    `json:"type"` // email, push, reminder
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

type EmailConfig struct {
	From     string
	To       string
	Subject  string
	Body     string
	Host     string
	Port     string
	Username string
	Password string
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

func (s *Service) CreateNotification(req NotificationRequest) (*models.Notification, error) {
	notification := models.Notification{
		UserID:      req.UserID,
		Title:       req.Title,
		Message:     req.Message,
		Type:        req.Type,
		IsRead:      false,
		ScheduledAt: req.ScheduledAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(&notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// If it's an email notification, send it immediately
	if req.Type == "email" && req.ScheduledAt == nil {
		go s.sendEmailNotification(notification)
	}

	// If it's scheduled, schedule it
	if req.ScheduledAt != nil {
		go s.scheduleNotification(notification)
	}

	return &notification, nil
}

func (s *Service) GetNotifications(userID uint, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	return notifications, nil
}

func (s *Service) GetUnreadNotifications(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.Where("user_id = ? AND is_read = false", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unread notifications: %w", err)
	}

	return notifications, nil
}

func (s *Service) MarkAsRead(notificationID uint, userID uint) error {
	var notification models.Notification
	if err := s.db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	notification.IsRead = true
	notification.UpdatedAt = time.Now()

	if err := s.db.Save(&notification).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

func (s *Service) MarkAllAsRead(userID uint) error {
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).
		Update("updated_at", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

func (s *Service) DeleteNotification(notificationID uint, userID uint) error {
	var notification models.Notification
	if err := s.db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	if err := s.db.Delete(&notification).Error; err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

func (s *Service) sendEmailNotification(notification models.Notification) error {
	// Get user email
	var user models.User
	if err := s.db.First(&user, notification.UserID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Email configuration (would normally come from config)
	emailConfig := EmailConfig{
		From:     "noreply@herlin.ai",
		To:       user.Email,
		Subject:  notification.Title,
		Body:     notification.Message,
		Host:     "smtp.gmail.com",
		Port:     "587",
		Username: s.cfg.Email.Username,
		Password: s.cfg.Email.Password,
	}

	return s.sendEmail(emailConfig)
}

func (s *Service) sendEmail(config EmailConfig) error {
	// Check if email credentials are configured
	if config.Username == "" || config.Password == "" {
		return fmt.Errorf("email credentials not configured")
	}

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		config.From, config.To, config.Subject, config.Body)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", config.Host, config.Port),
		auth,
		config.From,
		[]string{config.To},
		[]byte(message),
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *Service) scheduleNotification(notification models.Notification) {
	now := time.Now()
	if notification.ScheduledAt == nil || notification.ScheduledAt.Before(now) {
		return
	}

	duration := notification.ScheduledAt.Sub(now)
	time.AfterFunc(duration, func() {
		if notification.Type == "email" {
			s.sendEmailNotification(notification)
		}
		// Handle other notification types (push, etc.)
	})
}

func (s *Service) CreateReminder(userID uint, title string, message string, scheduledAt time.Time) (*models.Notification, error) {
	req := NotificationRequest{
		UserID:      userID,
		Title:       title,
		Message:     message,
		Type:        "reminder",
		ScheduledAt: &scheduledAt,
	}

	return s.CreateNotification(req)
}

func (s *Service) SendPushNotification(userID uint, title string, message string) error {
	// This would integrate with push notification services like Firebase Cloud Messaging
	// For now, just create a notification record
	req := NotificationRequest{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    "push",
	}

	_, err := s.CreateNotification(req)
	return err
}
