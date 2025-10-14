package services

import (
	"context"
	"fmt"

	"backend-residuos-app/pkg/firebase"

	"firebase.google.com/go/v4/messaging"
)

type NotificationService struct {
	firebaseClient *firebase.Firebase
}

// NewNotificationService crea una nueva instancia del servicio de notificaciones
func NewNotificationService(firebaseClient *firebase.Firebase) *NotificationService {
	return &NotificationService{
		firebaseClient: firebaseClient,
	}
}

// NotificationData estructura para los datos de notificación
type NotificationData struct {
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	ImageURL string            `json:"image_url,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

// SendToToken envía una notificación a un token específico
func (ns *NotificationService) SendToToken(ctx context.Context, token string, notification NotificationData) (string, error) {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		},
		Data: notification.Data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Priority: messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: notification.Title,
						Body:  notification.Body,
					},
					Badge: &[]int{1}[0],
					Sound: "default",
				},
			},
		},
	}

	response, err := ns.firebaseClient.SendNotification(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send notification to token %s: %v", token, err)
	}

	return response, nil
}

// SendToMultipleTokens envía una notificación a múltiples tokens
func (ns *NotificationService) SendToMultipleTokens(ctx context.Context, tokens []string, notification NotificationData) (*messaging.BatchResponse, error) {
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		},
		Data: notification.Data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Priority: messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: notification.Title,
						Body:  notification.Body,
					},
					Badge: &[]int{1}[0],
					Sound: "default",
				},
			},
		},
	}

	response, err := ns.firebaseClient.SendMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send multicast notification: %v", err)
	}

	return response, nil
}

// SendToTopic envía una notificación a un tema
func (ns *NotificationService) SendToTopic(ctx context.Context, topic string, notification NotificationData) (string, error) {
	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		},
		Data: notification.Data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Priority: messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: notification.Title,
						Body:  notification.Body,
					},
					Badge: &[]int{1}[0],
					Sound: "default",
				},
			},
		},
	}

	response, err := ns.firebaseClient.SendNotification(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send notification to topic %s: %v", topic, err)
	}

	return response, nil
}

// SubscribeToTopic suscribe tokens a un tema
func (ns *NotificationService) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	response, err := ns.firebaseClient.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("failed to subscribe tokens to topic %s: %v", topic, err)
	}

	// Verificar si hubo errores en tokens individuales
	if response.FailureCount > 0 {
		return fmt.Errorf("failed to subscribe %d tokens to topic %s", response.FailureCount, topic)
	}

	return nil
}

// UnsubscribeFromTopic desuscribe tokens de un tema
func (ns *NotificationService) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	response, err := ns.firebaseClient.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe tokens from topic %s: %v", topic, err)
	}

	// Verificar si hubo errores en tokens individuales
	if response.FailureCount > 0 {
		return fmt.Errorf("failed to unsubscribe %d tokens from topic %s", response.FailureCount, topic)
	}

	return nil
}

// SendWasteCollectionReminder envía una notificación de recordatorio de recolección de residuos
func (ns *NotificationService) SendWasteCollectionReminder(ctx context.Context, token string, collectionDate, wasteType string) (string, error) {
	notification := NotificationData{
		Title: "Recordatorio de Recolección",
		Body:  fmt.Sprintf("Recuerda sacar tus %s para la recolección del %s", wasteType, collectionDate),
		Data: map[string]string{
			"type":            "waste_collection_reminder",
			"collection_date": collectionDate,
			"waste_type":      wasteType,
		},
	}

	return ns.SendToToken(ctx, token, notification)
}

// SendReportStatusUpdate envía una notificación de actualización de estado de reporte
func (ns *NotificationService) SendReportStatusUpdate(ctx context.Context, token string, reportID, newStatus string) (string, error) {
	notification := NotificationData{
		Title: "Actualización de Reporte",
		Body:  fmt.Sprintf("Tu reporte #%s ha sido %s", reportID, newStatus),
		Data: map[string]string{
			"type":       "report_status_update",
			"report_id":  reportID,
			"new_status": newStatus,
		},
	}

	return ns.SendToToken(ctx, token, notification)
}
