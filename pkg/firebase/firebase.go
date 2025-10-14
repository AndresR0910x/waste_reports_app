package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Firebase struct {
	App       *firebase.App
	Auth      *auth.Client
	Messaging *messaging.Client
}

// InitializeFirebase inicializa Firebase con las credenciales y proyecto ID
func InitializeFirebase(credentialsPath, projectID string) (*Firebase, error) {
	ctx := context.Background()

	var opts []option.ClientOption

	// Si se proporciona un archivo de credenciales, usarlo
	if credentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
	}

	// Configurar el proyecto ID si se proporciona
	config := &firebase.Config{
		ProjectID: projectID,
	}

	// Inicializar la aplicación Firebase
	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	// Inicializar el cliente de autenticación
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase Auth: %v", err)
	}

	// Inicializar el cliente de mensajería (FCM)
	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase Messaging: %v", err)
	}

	return &Firebase{
		App:       app,
		Auth:      authClient,
		Messaging: messagingClient,
	}, nil
}

// VerifyIDToken verifica un token de ID de Firebase
func (f *Firebase) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := f.Auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying ID token: %v", err)
	}
	return token, nil
}

// GetUser obtiene un usuario por UID
func (f *Firebase) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	user, err := f.Auth.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	return user, nil
}

// CreateUser crea un nuevo usuario
func (f *Firebase) CreateUser(ctx context.Context, params *auth.UserToCreate) (*auth.UserRecord, error) {
	user, err := f.Auth.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %v", err)
	}
	return user, nil
}

// UpdateUser actualiza un usuario existente
func (f *Firebase) UpdateUser(ctx context.Context, uid string, params *auth.UserToUpdate) (*auth.UserRecord, error) {
	user, err := f.Auth.UpdateUser(ctx, uid, params)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %v", err)
	}
	return user, nil
}

// DeleteUser elimina un usuario
func (f *Firebase) DeleteUser(ctx context.Context, uid string) error {
	err := f.Auth.DeleteUser(ctx, uid)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}
	return nil
}

// SendNotification envía una notificación push usando FCM
func (f *Firebase) SendNotification(ctx context.Context, message *messaging.Message) (string, error) {
	response, err := f.Messaging.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("error sending message: %v", err)
	}
	return response, nil
}

// SendMulticast envía notificaciones a múltiples tokens
func (f *Firebase) SendMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	response, err := f.Messaging.SendMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("error sending multicast message: %v", err)
	}
	return response, nil
}

// SubscribeToTopic suscribe tokens a un tema
func (f *Firebase) SubscribeToTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	response, err := f.Messaging.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return nil, fmt.Errorf("error subscribing to topic: %v", err)
	}
	return response, nil
}

// UnsubscribeFromTopic desuscribe tokens de un tema
func (f *Firebase) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	response, err := f.Messaging.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return nil, fmt.Errorf("error unsubscribing from topic: %v", err)
	}
	return response, nil
}
