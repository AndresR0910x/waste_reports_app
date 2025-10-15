package tests

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"backend-residuos-app/internal/services"
	"backend-residuos-app/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// PhotoUploadServiceTestSuite define la suite de tests
type PhotoUploadServiceTestSuite struct {
	suite.Suite
	service       *services.PhotoUploadService
	mockStorage   *mocks.MockStorageClient
	mockBucket    *mocks.MockBucketHandle
	mockObject    *mocks.MockObjectHandle
	mockWriter    *mocks.MockWriter
	mockLogger    *mocks.MockLogger
	bucketName    string
	publicBaseURL string
}

// SetupTest configura cada test
func (suite *PhotoUploadServiceTestSuite) SetupTest() {
	suite.bucketName = "test-bucket"
	suite.publicBaseURL = "https://storage.googleapis.com/" + suite.bucketName + "/"

	suite.mockStorage = new(mocks.MockStorageClient)
	suite.mockBucket = new(mocks.MockBucketHandle)
	suite.mockObject = new(mocks.MockObjectHandle)
	suite.mockWriter = new(mocks.MockWriter)
	suite.mockLogger = new(mocks.MockLogger)

	suite.service = services.NewPhotoUploadService(
		suite.mockStorage,
		suite.bucketName,
		suite.publicBaseURL,
		suite.mockLogger,
	)
}

// TestUploadFromBase64_Success prueba subida exitosa desde base64
func (suite *PhotoUploadServiceTestSuite) TestUploadFromBase64_Success() {
	ctx := context.Background()

	// Crear imagen base64 simple (1x1 pixel PNG)
	pngData := "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\tpHYs\x00\x00\x0b\x13\x00\x00\x0b\x13\x01\x00\x9a\x9c\x18\x00\x00\x00\nIDATx\x9cc\xf8\x00\x00\x00\x01\x00\x01\x00\x00\x00\x00\x37\x82\xfc\x0b\x00\x00\x00\x00IEND\xaeB`\x82"
	base64Data := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(pngData))
	baseFilename := "test-image"

	// Configurar mocks
	suite.mockStorage.On("Bucket", suite.bucketName).Return(suite.mockBucket)
	suite.mockBucket.On("Object", mock.MatchedBy(func(name string) bool {
		return strings.Contains(name, "test-image") && strings.HasSuffix(name, ".png")
	})).Return(suite.mockObject)

	suite.mockObject.On("NewWriter", ctx).Return(suite.mockWriter)
	suite.mockWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(len(pngData), nil)
	suite.mockWriter.On("Close").Return(nil)

	suite.mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	result, err := suite.service.UploadFromBase64(ctx, base64Data, baseFilename)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), result.Filename, "test-image")
	assert.Contains(suite.T(), result.Filename, ".png")
	assert.Contains(suite.T(), result.URL, suite.publicBaseURL)
	assert.Equal(suite.T(), "image/png", result.ContentType)
	assert.Greater(suite.T(), result.Size, int64(0))

	// Verificar que se llamaron los mocks
	suite.mockStorage.AssertExpectations(suite.T())
	suite.mockBucket.AssertExpectations(suite.T())
	suite.mockObject.AssertExpectations(suite.T())
	suite.mockWriter.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestUploadFromBase64_InvalidFormat prueba formato base64 inválido
func (suite *PhotoUploadServiceTestSuite) TestUploadFromBase64_InvalidFormat() {
	ctx := context.Background()
	invalidBase64 := "invalid-base64-data"

	suite.mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	result, err := suite.service.UploadFromBase64(ctx, invalidBase64, "test")

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "formato base64 inválido")

	suite.mockLogger.AssertExpectations(suite.T())
}

// TestGenerateFileName prueba generación de nombres de archivo
func (suite *PhotoUploadServiceTestSuite) TestGenerateFileName() {
	baseName := "test-image"
	extension := ".jpg"

	// Ejecutar función
	result := suite.service.GenerateFileName(baseName, extension)

	// Verificar resultado
	assert.Contains(suite.T(), result, baseName)
	assert.Contains(suite.T(), result, extension)
	assert.Greater(suite.T(), len(result), len(baseName+extension))
}

// TestGetFileExtensionFromMimeType prueba obtención de extensión
func (suite *PhotoUploadServiceTestSuite) TestGetFileExtensionFromMimeType() {
	testCases := []struct {
		mimeType string
		expected string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"unknown/type", ".bin"},
	}

	for _, tc := range testCases {
		result := suite.service.GetFileExtensionFromMimeType(tc.mimeType)
		assert.Equal(suite.T(), tc.expected, result, "MimeType: %s", tc.mimeType)
	}
}

// TestIsSupportedImageType prueba validación de tipos soportados
func (suite *PhotoUploadServiceTestSuite) TestIsSupportedImageType() {
	testCases := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"text/plain", false},
		{"application/pdf", false},
		{"video/mp4", false},
	}

	for _, tc := range testCases {
		result := suite.service.IsSupportedImageType(tc.mimeType)
		assert.Equal(suite.T(), tc.expected, result, "MimeType: %s", tc.mimeType)
	}
}

// TestPhotoUploadService ejecuta toda la suite
func TestPhotoUploadService(t *testing.T) {
	suite.Run(t, new(PhotoUploadServiceTestSuite))
}
