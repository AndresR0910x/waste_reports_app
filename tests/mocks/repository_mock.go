package mocks

import (
	"backend-residuos-app/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockReportRepository mock del repositorio de reportes
type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) Create(report *models.Report) (*models.Report, error) {
	args := m.Called(report)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Report), args.Error(1)
}

func (m *MockReportRepository) GetByID(id int) (*models.Report, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Report), args.Error(1)
}

func (m *MockReportRepository) GetByUserID(userID, limit, offset int) ([]*models.Report, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportRepository) Update(report *models.Report) error {
	args := m.Called(report)
	return args.Error(0)
}

func (m *MockReportRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockReportRepository) GetByStatus(status string, limit, offset int) ([]*models.Report, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportRepository) GetByCategory(category string, limit, offset int) ([]*models.Report, error) {
	args := m.Called(category, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportRepository) GetByLocation(lat, lng, radius float64, limit, offset int) ([]*models.Report, error) {
	args := m.Called(lat, lng, radius, limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportRepository) CountByUserID(userID int) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockReportRepository) GetAll(limit, offset int) ([]*models.Report, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportRepository) CountByStatus(status string) (int, error) {
	args := m.Called(status)
	return args.Int(0), args.Error(1)
}

func (m *MockReportRepository) CountByCategory(category string) (int, error) {
	args := m.Called(category)
	return args.Int(0), args.Error(1)
}

func (m *MockReportRepository) GetRecentReports(limit int) ([]*models.Report, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.Report), args.Error(1)
}

func (m *MockReportRepository) UpdateSyncStatus(id int, syncStatus string) error {
	args := m.Called(id, syncStatus)
	return args.Error(0)
}

func (m *MockReportRepository) GetUnsyncedReports() ([]*models.Report, error) {
	args := m.Called()
	return args.Get(0).([]*models.Report), args.Error(1)
}

// MockLogger mock del logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(message string, fields ...interface{}) {
	args := []interface{}{message}
	args = append(args, fields...)
	m.Called(args...)
}

func (m *MockLogger) Error(message string, fields ...interface{}) {
	args := []interface{}{message}
	args = append(args, fields...)
	m.Called(args...)
}

func (m *MockLogger) Warn(message string, fields ...interface{}) {
	args := []interface{}{message}
	args = append(args, fields...)
	m.Called(args...)
}

func (m *MockLogger) Debug(message string, fields ...interface{}) {
	args := []interface{}{message}
	args = append(args, fields...)
	m.Called(args...)
}

func (m *MockLogger) Fatal(message string, fields ...interface{}) {
	args := []interface{}{message}
	args = append(args, fields...)
	m.Called(args...)
}
