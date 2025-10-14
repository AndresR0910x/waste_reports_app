package handlers

import (
	"net/http"

	"backend-residuos-app/internal/services"
	"backend-residuos-app/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

// NewNotificationHandler crea una nueva instancia del handler de notificaciones
func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendNotificationRequest estructura para el request de envío de notificación
type SendNotificationRequest struct {
	Token    string            `json:"token" binding:"required"`
	Title    string            `json:"title" binding:"required"`
	Body     string            `json:"body" binding:"required"`
	ImageURL string            `json:"image_url,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

// SendMulticastRequest estructura para el request de envío múltiple
type SendMulticastRequest struct {
	Tokens   []string          `json:"tokens" binding:"required"`
	Title    string            `json:"title" binding:"required"`
	Body     string            `json:"body" binding:"required"`
	ImageURL string            `json:"image_url,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

// SendToTopicRequest estructura para el request de envío a tema
type SendToTopicRequest struct {
	Topic    string            `json:"topic" binding:"required"`
	Title    string            `json:"title" binding:"required"`
	Body     string            `json:"body" binding:"required"`
	ImageURL string            `json:"image_url,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

// SubscribeToTopicRequest estructura para suscripción a tema
type SubscribeToTopicRequest struct {
	Tokens []string `json:"tokens" binding:"required"`
	Topic  string   `json:"topic" binding:"required"`
}

// SendNotification envía una notificación a un token específico
func (nh *NotificationHandler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	notification := services.NotificationData{
		Title:    req.Title,
		Body:     req.Body,
		ImageURL: req.ImageURL,
		Data:     req.Data,
	}

	response, err := nh.notificationService.SendToToken(c.Request.Context(), req.Token, notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send notification",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Notification sent successfully",
		"message_id": response,
	})
}

// SendMulticast envía notificaciones a múltiples tokens
func (nh *NotificationHandler) SendMulticast(c *gin.Context) {
	var req SendMulticastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	notification := services.NotificationData{
		Title:    req.Title,
		Body:     req.Body,
		ImageURL: req.ImageURL,
		Data:     req.Data,
	}

	response, err := nh.notificationService.SendToMultipleTokens(c.Request.Context(), req.Tokens, notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send multicast notification",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Multicast notification sent successfully",
		"success_count": response.SuccessCount,
		"failure_count": response.FailureCount,
	})
}

// SendToTopic envía una notificación a un tema
func (nh *NotificationHandler) SendToTopic(c *gin.Context) {
	var req SendToTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	notification := services.NotificationData{
		Title:    req.Title,
		Body:     req.Body,
		ImageURL: req.ImageURL,
		Data:     req.Data,
	}

	response, err := nh.notificationService.SendToTopic(c.Request.Context(), req.Topic, notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send topic notification",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Topic notification sent successfully",
		"message_id": response,
	})
}

// SubscribeToTopic suscribe tokens a un tema
func (nh *NotificationHandler) SubscribeToTopic(c *gin.Context) {
	var req SubscribeToTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	err := nh.notificationService.SubscribeToTopic(c.Request.Context(), req.Tokens, req.Topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to subscribe to topic",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully subscribed to topic",
	})
}

// UnsubscribeFromTopic desuscribe tokens de un tema
func (nh *NotificationHandler) UnsubscribeFromTopic(c *gin.Context) {
	var req SubscribeToTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	err := nh.notificationService.UnsubscribeFromTopic(c.Request.Context(), req.Tokens, req.Topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to unsubscribe from topic",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully unsubscribed from topic",
	})
}

// SendWasteCollectionReminder envía recordatorio de recolección de residuos
func (nh *NotificationHandler) SendWasteCollectionReminder(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	type ReminderRequest struct {
		Token          string `json:"token" binding:"required"`
		CollectionDate string `json:"collection_date" binding:"required"`
		WasteType      string `json:"waste_type" binding:"required"`
	}

	var req ReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	response, err := nh.notificationService.SendWasteCollectionReminder(
		c.Request.Context(),
		req.Token,
		req.CollectionDate,
		req.WasteType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send reminder",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Waste collection reminder sent successfully",
		"message_id": response,
		"user_id":    userID,
	})
}
