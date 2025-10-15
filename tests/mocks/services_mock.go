package mocks

import (
	"context"
	"io"

	"backend-residuos-app/internal/models"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/mock"
)

// MockReportService mock del servicio de reportes
type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) CreateReport(report *models.Report) (*models.Report, error) {
	args := m.Called(report)
	return args.Get(0).(*models.Report), args.Error(1)
}

func (m *MockReportService) GetReportByID(id int) (*models.Report, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Report), args.Error(1)
}

func (m *MockReportService) GetReportsByUserID(userID, limit, offset int) ([]*models.Report, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportService) GetReportsCountByUserID(userID int) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockReportService) UpdateReport(report *models.Report) error {
	args := m.Called(report)
	return args.Error(0)
}

func (m *MockReportService) DeleteReport(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockReportService) GetReportsByStatus(status string, limit, offset int) ([]*models.Report, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportService) GetReportsByCategory(category string, limit, offset int) ([]*models.Report, error) {
	args := m.Called(category, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportService) GetReportsByLocation(lat, lng, radius float64, limit, offset int) ([]*models.Report, error) {
	args := m.Called(lat, lng, radius, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

// MockCedulaValidationService mock del servicio de validación de cédula
type MockCedulaValidationService struct {
	mock.Mock
}

func (m *MockCedulaValidationService) ValidateCedula(cedula string) (*models.CedulaValidationResponse, error) {
	args := m.Called(cedula)
	return args.Get(0).(*models.CedulaValidationResponse), args.Error(1)
}

func (m *MockCedulaValidationService) IsValidCedulaFormat(cedula string) bool {
	args := m.Called(cedula)
	return args.Bool(0)
}

func (m *MockCedulaValidationService) CleanCedula(cedula string) string {
	args := m.Called(cedula)
	return args.String(0)
}

func (m *MockCedulaValidationService) ValidateWithAlgorithm(cedula string) (*models.CedulaValidationResponse, error) {
	args := m.Called(cedula)
	return args.Get(0).(*models.CedulaValidationResponse), args.Error(1)
}

func (m *MockCedulaValidationService) ValidateWithAPI(cedula string) (*models.CedulaValidationResponse, error) {
	args := m.Called(cedula)
	return args.Get(0).(*models.CedulaValidationResponse), args.Error(1)
}

func (m *MockCedulaValidationService) GetProvinceName(provinceCode int) string {
	args := m.Called(provinceCode)
	return args.String(0)
}

// MockPhotoUploadService mock del servicio de subida de fotos
type MockPhotoUploadService struct {
	mock.Mock
}

func (m *MockPhotoUploadService) UploadFromBase64(ctx context.Context, base64Data string, baseFilename string) (*models.PhotoUploadResult, error) {
	args := m.Called(ctx, base64Data, baseFilename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PhotoUploadResult), args.Error(1)
}

func (m *MockPhotoUploadService) UploadFromStream(ctx context.Context, reader io.Reader, contentType string, baseFilename string) (*models.PhotoUploadResult, error) {
	args := m.Called(ctx, reader, contentType, baseFilename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PhotoUploadResult), args.Error(1)
}

func (m *MockPhotoUploadService) DeletePhoto(ctx context.Context, filename string) error {
	args := m.Called(ctx, filename)
	return args.Error(0)
}

func (m *MockPhotoUploadService) GenerateFileName(baseName string, extension string) string {
	args := m.Called(baseName, extension)
	return args.String(0)
}

func (m *MockPhotoUploadService) GetFileExtensionFromMimeType(mimeType string) string {
	args := m.Called(mimeType)
	return args.String(0)
}

func (m *MockPhotoUploadService) IsSupportedImageType(mimeType string) bool {
	args := m.Called(mimeType)
	return args.Bool(0)
}

// MockValidationService mock del servicio de validación
type MockValidationService struct {
	mock.Mock
}

func (m *MockValidationService) ValidateStruct(s interface{}) []models.ValidationError {
	args := m.Called(s)
	return args.Get(0).([]models.ValidationError)
}

func (m *MockValidationService) ValidateField(field string, value interface{}, rule string) error {
	args := m.Called(field, value, rule)
	return args.Error(0)
}

// MockStorageClient mock del cliente de Google Cloud Storage
type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) Bucket(name string) *storage.BucketHandle {
	args := m.Called(name)
	return args.Get(0).(*storage.BucketHandle)
}

// MockBucketHandle mock del handle del bucket
type MockBucketHandle struct {
	mock.Mock
}

func (m *MockBucketHandle) Object(name string) *storage.ObjectHandle {
	args := m.Called(name)
	return args.Get(0).(*storage.ObjectHandle)
}

// MockObjectHandle mock del handle del objeto
type MockObjectHandle struct {
	mock.Mock
}

func (m *MockObjectHandle) NewWriter(ctx context.Context) *storage.Writer {
	args := m.Called(ctx)
	return args.Get(0).(*storage.Writer)
}

func (m *MockObjectHandle) Delete(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockObjectHandle) Attrs(ctx context.Context) (*storage.ObjectAttrs, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.ObjectAttrs), args.Error(1)
}

// MockWriter mock del writer de Google Cloud Storage
type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}
