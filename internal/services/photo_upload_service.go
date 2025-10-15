package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// PhotoUploadService maneja la subida de fotos a Firebase Storage
type PhotoUploadService struct {
	client     *storage.Client
	bucketName string
}

// NewPhotoUploadService crea una nueva instancia del servicio
func NewPhotoUploadService(bucketName string, credentialsPath string) (*PhotoUploadService, error) {
	ctx := context.Background()

	var client *storage.Client
	var err error

	if credentialsPath != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	} else {
		// Usar credenciales por defecto (variables de entorno)
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &PhotoUploadService{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadPhotoFromBase64 sube una foto desde base64 a Firebase Storage
func (s *PhotoUploadService) UploadPhotoFromBase64(base64Data, fileName string, userID int) (string, error) {
	if base64Data == "" {
		return "", fmt.Errorf("base64 data is empty")
	}

	// Detectar y remover el prefijo data:image/...;base64,
	base64Data = s.cleanBase64Data(base64Data)

	// Decodificar base64
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Generar nombre único para el archivo
	uniqueFileName := s.generateUniqueFileName(fileName, userID)

	// Subir a Firebase Storage
	return s.uploadToStorage(imageData, uniqueFileName)
}

// UploadPhoto sube una foto desde un reader a Firebase Storage
func (s *PhotoUploadService) UploadPhoto(reader io.Reader, fileName string, userID int) (string, error) {
	// Generar nombre único para el archivo
	uniqueFileName := s.generateUniqueFileName(fileName, userID)

	// Leer datos del reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read photo data: %w", err)
	}

	// Subir a Firebase Storage
	return s.uploadToStorage(data, uniqueFileName)
}

// uploadToStorage realiza la subida real a Firebase Storage
func (s *PhotoUploadService) uploadToStorage(data []byte, fileName string) (string, error) {
	ctx := context.Background()

	// Obtener referencia al bucket
	bucket := s.client.Bucket(s.bucketName)

	// Crear object handle
	obj := bucket.Object(fileName)

	// Crear writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = s.detectContentType(data)
	writer.Metadata = map[string]string{
		"uploadedAt": time.Now().Format(time.RFC3339),
		"source":     "clean-city-app",
	}

	// Escribir datos
	if _, err := writer.Write(data); err != nil {
		return "", fmt.Errorf("failed to write to storage: %w", err)
	}

	// Cerrar writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Hacer el archivo público (opcional)
	if err := s.makeObjectPublic(ctx, obj); err != nil {
		// Log error but don't fail the upload
		fmt.Printf("Warning: failed to make object public: %v\n", err)
	}

	// Construir URL de descarga
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, fileName)

	return publicURL, nil
}

// makeObjectPublic hace que un objeto sea público
func (s *PhotoUploadService) makeObjectPublic(ctx context.Context, obj *storage.ObjectHandle) error {
	acl := obj.ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return fmt.Errorf("failed to set object ACL: %w", err)
	}
	return nil
}

// DeletePhoto elimina una foto de Firebase Storage
func (s *PhotoUploadService) DeletePhoto(photoURL string) error {
	ctx := context.Background()

	// Extraer el nombre del archivo de la URL
	fileName := s.extractFileNameFromURL(photoURL)
	if fileName == "" {
		return fmt.Errorf("cannot extract filename from URL: %s", photoURL)
	}

	// Obtener referencia al object
	obj := s.client.Bucket(s.bucketName).Object(fileName)

	// Eliminar el objeto
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	return nil
}

// cleanBase64Data limpia los datos base64 removiendo prefijos
func (s *PhotoUploadService) cleanBase64Data(base64Data string) string {
	// Buscar el índice de la coma que separa el prefijo de los datos
	if idx := strings.Index(base64Data, ","); idx != -1 {
		return base64Data[idx+1:]
	}
	return base64Data
}

// generateUniqueFileName genera un nombre único para el archivo
func (s *PhotoUploadService) generateUniqueFileName(originalName string, userID int) string {
	timestamp := time.Now().Unix()

	// Obtener extensión del archivo original
	extension := ""
	if idx := strings.LastIndex(originalName, "."); idx != -1 {
		extension = originalName[idx:]
	} else {
		// Por defecto usar .jpg si no hay extensión
		extension = ".jpg"
	}

	// Construir nombre único: reports/user_123/1697361234_photo.jpg
	return fmt.Sprintf("reports/user_%d/%d_photo%s", userID, timestamp, extension)
}

// detectContentType detecta el tipo de contenido basado en los primeros bytes
func (s *PhotoUploadService) detectContentType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// Detectar tipo de imagen por magic numbers
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "image/jpeg"
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
		return "image/png"
	case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
		return "image/gif"
	case data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46:
		return "image/webp"
	default:
		return "image/jpeg" // Default to JPEG
	}
}

// extractFileNameFromURL extrae el nombre del archivo de una URL de Firebase Storage
func (s *PhotoUploadService) extractFileNameFromURL(url string) string {
	// Formato esperado: https://storage.googleapis.com/bucket-name/path/to/file.jpg
	prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", s.bucketName)

	if strings.HasPrefix(url, prefix) {
		return url[len(prefix):]
	}

	// Alternativa: si la URL tiene formato diferente, extraer solo el nombre del archivo
	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return url[idx+1:]
	}

	return ""
}

// ValidateImageSize valida el tamaño de la imagen
func (s *PhotoUploadService) ValidateImageSize(data []byte, maxSizeMB int) error {
	sizeMB := len(data) / (1024 * 1024)
	if sizeMB > maxSizeMB {
		return fmt.Errorf("image size (%d MB) exceeds maximum allowed size (%d MB)", sizeMB, maxSizeMB)
	}
	return nil
}

// Close cierra el cliente de storage
func (s *PhotoUploadService) Close() error {
	return s.client.Close()
}
