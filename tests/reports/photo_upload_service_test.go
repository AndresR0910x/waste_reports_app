package tests

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"testing"

	"backend-residuos-app/internal/services"
	"backend-residuos-app/tests/mocks"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// PhotoUploadServiceTestSuite define la suite de tests
type PhotoUploadServiceTestSuite struct {
	suite.Suite
	service    *services.PhotoUploadService
	mockClient *mocks.MockStorageClient
	mockBucket *mocks.MockBucketHandle
	mockObject *mocks.MockObjectHandle
	mockWriter *mocks.MockWriter
}

// SetupTest configura cada test
func (suite *PhotoUploadServiceTestSuite) SetupTest() {
	suite.mockClient = new(mocks.MockStorageClient)
	suite.mockBucket = new(mocks.MockBucketHandle)
	suite.mockObject = new(mocks.MockObjectHandle)
	suite.mockWriter = new(mocks.MockWriter)

	suite.service = services.NewPhotoUploadService(suite.mockClient, "test-bucket")
}

// TestUploadFromBase64_Success prueba la subida exitosa desde base64
func (suite *PhotoUploadServiceTestSuite) TestUploadFromBase64_Success() {
	// Crear imagen de prueba
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	assert.NoError(suite.T(), err)

	base64Data := "data:image/jpeg;base64," + "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="

	// Configurar mocks
	suite.mockClient.On("Bucket", "test-bucket").Return(suite.mockBucket)
	suite.mockBucket.On("Object", mock.MatchedBy(func(filename string) bool {
		return len(filename) > 10 // Verificar que se genere un nombre de archivo
	})).Return(suite.mockObject)

	suite.mockObject.On("NewWriter", mock.AnythingOfType("context.Context")).Return(suite.mockWriter)
	suite.mockWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(len(buf.Bytes()), nil)
	suite.mockWriter.On("Close").Return(nil)

	expectedURL := "https://storage.googleapis.com/test-bucket/test-image.jpg"
	suite.mockObject.On("Attrs", mock.AnythingOfType("context.Context")).Return(&storage.ObjectAttrs{
		MediaLink: expectedURL,
	}, nil)

	// Ejecutar función
	ctx := context.Background()
	result, err := suite.service.UploadFromBase64(ctx, base64Data, "test-image")

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result.URL)
	assert.NotEmpty(suite.T(), result.Filename)
	assert.Equal(suite.T(), "image/jpeg", result.ContentType)
	assert.Greater(suite.T(), result.Size, int64(0))

	// Verificar que se llamaron los mocks
	suite.mockClient.AssertExpectations(suite.T())
	suite.mockBucket.AssertExpectations(suite.T())
	suite.mockObject.AssertExpectations(suite.T())
	suite.mockWriter.AssertExpectations(suite.T())
}

// TestUploadFromBase64_InvalidFormat prueba formato base64 inválido
func (suite *PhotoUploadServiceTestSuite) TestUploadFromBase64_InvalidFormat() {
	base64Data := "invalid-base64-data"

	ctx := context.Background()
	result, err := suite.service.UploadFromBase64(ctx, base64Data, "test-image")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "formato de base64 inválido")
}

// TestUploadFromBase64_UnsupportedMimeType prueba tipo MIME no soportado
func (suite *PhotoUploadServiceTestSuite) TestUploadFromBase64_UnsupportedMimeType() {
	// Base64 de un archivo de texto
	base64Data := "data:text/plain;base64,SGVsbG8gV29ybGQ="

	ctx := context.Background()
	result, err := suite.service.UploadFromBase64(ctx, base64Data, "test-file")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "tipo de archivo no soportado")
}

// TestUploadFromBase64_FileSizeExceeded prueba límite de tamaño excedido
func (suite *PhotoUploadServiceTestSuite) TestUploadFromBase64_FileSizeExceeded() {
	// Crear una imagen grande (simulada con datos grandes)
	largeData := make([]byte, 11*1024*1024) // 11MB
	for i := range largeData {
		largeData[i] = 255
	}

	// Este es un base64 simplificado para el test
	base64Data := "data:image/jpeg;base64," + string(largeData)

	ctx := context.Background()
	result, err := suite.service.UploadFromBase64(ctx, base64Data, "test-image")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "archivo demasiado grande")
}

// TestUploadFromStream_Success prueba la subida exitosa desde stream
func (suite *PhotoUploadServiceTestSuite) TestUploadFromStream_Success() {
	// Crear imagen de prueba
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	assert.NoError(suite.T(), err)

	reader := bytes.NewReader(buf.Bytes())

	// Configurar mocks
	suite.mockClient.On("Bucket", "test-bucket").Return(suite.mockBucket)
	suite.mockBucket.On("Object", mock.MatchedBy(func(filename string) bool {
		return len(filename) > 10
	})).Return(suite.mockObject)

	suite.mockObject.On("NewWriter", mock.AnythingOfType("context.Context")).Return(suite.mockWriter)
	suite.mockWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(len(buf.Bytes()), nil)
	suite.mockWriter.On("Close").Return(nil)

	expectedURL := "https://storage.googleapis.com/test-bucket/test-image.jpg"
	suite.mockObject.On("Attrs", mock.AnythingOfType("context.Context")).Return(&storage.ObjectAttrs{
		MediaLink: expectedURL,
	}, nil)

	// Ejecutar función
	ctx := context.Background()
	result, err := suite.service.UploadFromStream(ctx, reader, "image/jpeg", "test-image")

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), result.URL)
	assert.NotEmpty(suite.T(), result.Filename)
	assert.Equal(suite.T(), "image/jpeg", result.ContentType)
	assert.Greater(suite.T(), result.Size, int64(0))

	suite.mockClient.AssertExpectations(suite.T())
	suite.mockBucket.AssertExpectations(suite.T())
	suite.mockObject.AssertExpectations(suite.T())
	suite.mockWriter.AssertExpectations(suite.T())
}

// TestUploadFromStream_WriteError prueba error durante la escritura
func (suite *PhotoUploadServiceTestSuite) TestUploadFromStream_WriteError() {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	assert.NoError(suite.T(), err)

	reader := bytes.NewReader(buf.Bytes())

	// Configurar mocks con error
	suite.mockClient.On("Bucket", "test-bucket").Return(suite.mockBucket)
	suite.mockBucket.On("Object", mock.AnythingOfType("string")).Return(suite.mockObject)
	suite.mockObject.On("NewWriter", mock.AnythingOfType("context.Context")).Return(suite.mockWriter)
	suite.mockWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(0, errors.New("write error"))
	suite.mockWriter.On("Close").Return(nil)

	// Ejecutar función
	ctx := context.Background()
	result, err := suite.service.UploadFromStream(ctx, reader, "image/jpeg", "test-image")

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "error al subir archivo")

	suite.mockClient.AssertExpectations(suite.T())
	suite.mockBucket.AssertExpectations(suite.T())
	suite.mockObject.AssertExpectations(suite.T())
	suite.mockWriter.AssertExpectations(suite.T())
}

// TestGenerateFileName prueba la generación de nombres de archivo
func (suite *PhotoUploadServiceTestSuite) TestGenerateFileName() {
	filename1 := suite.service.GenerateFileName("test", "jpg")
	filename2 := suite.service.GenerateFileName("test", "jpg")

	// Los nombres deben ser diferentes (incluyen timestamp)
	assert.NotEqual(suite.T(), filename1, filename2)
	assert.Contains(suite.T(), filename1, "test")
	assert.Contains(suite.T(), filename1, ".jpg")
	assert.Contains(suite.T(), filename2, "test")
	assert.Contains(suite.T(), filename2, ".jpg")
}

// TestGetFileExtensionFromMimeType prueba obtener extensión desde MIME type
func (suite *PhotoUploadServiceTestSuite) TestGetFileExtensionFromMimeType() {
	tests := []struct {
		mimeType string
		expected string
	}{
		{"image/jpeg", "jpg"},
		{"image/jpg", "jpg"},
		{"image/png", "png"},
		{"image/gif", "gif"},
		{"image/webp", "webp"},
		{"unknown/type", "bin"},
	}

	for _, test := range tests {
		result := suite.service.GetFileExtensionFromMimeType(test.mimeType)
		assert.Equal(suite.T(), test.expected, result, "Para MIME type: %s", test.mimeType)
	}
}

// TestIsSupportedImageType prueba la validación de tipos de imagen soportados
func (suite *PhotoUploadServiceTestSuite) TestIsSupportedImageType() {
	supportedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	unsupportedTypes := []string{
		"text/plain",
		"application/pdf",
		"video/mp4",
		"audio/mp3",
	}

	for _, mimeType := range supportedTypes {
		assert.True(suite.T(), suite.service.IsSupportedImageType(mimeType), "Debe soportar: %s", mimeType)
	}

	for _, mimeType := range unsupportedTypes {
		assert.False(suite.T(), suite.service.IsSupportedImageType(mimeType), "No debe soportar: %s", mimeType)
	}
}

// TestDeletePhoto prueba la eliminación de fotos
func (suite *PhotoUploadServiceTestSuite) TestDeletePhoto_Success() {
	filename := "reports/test-photo-123.jpg"

	// Configurar mocks
	suite.mockClient.On("Bucket", "test-bucket").Return(suite.mockBucket)
	suite.mockBucket.On("Object", filename).Return(suite.mockObject)
	suite.mockObject.On("Delete", mock.AnythingOfType("context.Context")).Return(nil)

	// Ejecutar función
	ctx := context.Background()
	err := suite.service.DeletePhoto(ctx, filename)

	// Verificar resultado
	assert.NoError(suite.T(), err)

	suite.mockClient.AssertExpectations(suite.T())
	suite.mockBucket.AssertExpectations(suite.T())
	suite.mockObject.AssertExpectations(suite.T())
}

// TestDeletePhoto_Error prueba error al eliminar foto
func (suite *PhotoUploadServiceTestSuite) TestDeletePhoto_Error() {
	filename := "reports/test-photo-123.jpg"

	// Configurar mocks con error
	suite.mockClient.On("Bucket", "test-bucket").Return(suite.mockBucket)
	suite.mockBucket.On("Object", filename).Return(suite.mockObject)
	suite.mockObject.On("Delete", mock.AnythingOfType("context.Context")).Return(errors.New("delete error"))

	// Ejecutar función
	ctx := context.Background()
	err := suite.service.DeletePhoto(ctx, filename)

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "error al eliminar foto")

	suite.mockClient.AssertExpectations(suite.T())
	suite.mockBucket.AssertExpectations(suite.T())
	suite.mockObject.AssertExpectations(suite.T())
}

// TestPhotoUploadService ejecuta toda la suite
func TestPhotoUploadService(t *testing.T) {
	suite.Run(t, new(PhotoUploadServiceTestSuite))
}
