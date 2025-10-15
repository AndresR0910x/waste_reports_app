package tests

import (
	"errors"
	"testing"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/internal/services"
	"backend-residuos-app/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ReportServiceTestSuite define la suite de tests
type ReportServiceTestSuite struct {
	suite.Suite
	service    *services.ReportService
	mockRepo   *mocks.MockReportRepository
	mockLogger *mocks.MockLogger
}

// SetupTest configura cada test
func (suite *ReportServiceTestSuite) SetupTest() {
	suite.mockRepo = new(mocks.MockReportRepository)
	suite.mockLogger = new(mocks.MockLogger)

	suite.service = services.NewReportService(suite.mockRepo, suite.mockLogger)
}

// TestCreateReport_Success prueba la creación exitosa de un reporte
func (suite *ReportServiceTestSuite) TestCreateReport_Success() {
	// Preparar datos de entrada
	inputReport := &models.Report{
		UserID:      1,
		Title:       "Basura en la calle",
		Description: "Hay mucha basura acumulada",
		Category:    "limpieza_publica",
		Location: models.Location{
			Latitude:  -0.1807,
			Longitude: -78.4678,
		},
		Address:  stringPtr("Av. Amazonas y Naciones Unidas"),
		Priority: "media",
	}

	// Preparar respuesta esperada
	expectedReport := &models.Report{
		ID:          1,
		UserID:      1,
		Title:       "Basura en la calle",
		Description: "Hay mucha basura acumulada",
		Category:    "limpieza_publica",
		Location: models.Location{
			Latitude:  -0.1807,
			Longitude: -78.4678,
		},
		Address:    stringPtr("Av. Amazonas y Naciones Unidas"),
		Status:     "pending",
		Priority:   "media",
		SyncStatus: "synced",
	}

	// Configurar mocks
	suite.mockRepo.On("Create", mock.MatchedBy(func(report *models.Report) bool {
		return report.Status == "pending" && report.SyncStatus == "synced"
	})).Return(expectedReport, nil)

	suite.mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	result, err := suite.service.CreateReport(inputReport)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedReport.ID, result.ID)
	assert.Equal(suite.T(), expectedReport.Title, result.Title)
	assert.Equal(suite.T(), "pending", result.Status)
	assert.Equal(suite.T(), "synced", result.SyncStatus)

	// Verificar que se llamaron los mocks
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestCreateReport_RepositoryError prueba error del repositorio
func (suite *ReportServiceTestSuite) TestCreateReport_RepositoryError() {
	inputReport := &models.Report{
		UserID:      1,
		Title:       "Test Report",
		Description: "Test Description",
		Category:    "limpieza_publica",
	}

	// Configurar mocks con error
	suite.mockRepo.On("Create", mock.AnythingOfType("*models.Report")).
		Return(nil, errors.New("database connection failed"))

	suite.mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	result, err := suite.service.CreateReport(inputReport)

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "error al crear reporte")

	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestGetReportByID_Success prueba obtener reporte por ID exitosamente
func (suite *ReportServiceTestSuite) TestGetReportByID_Success() {
	expectedReport := &models.Report{
		ID:          1,
		UserID:      1,
		Title:       "Test Report",
		Description: "Test Description",
		Status:      "pending",
	}

	// Configurar mock
	suite.mockRepo.On("GetByID", 1).Return(expectedReport, nil)

	// Ejecutar función
	result, err := suite.service.GetReportByID(1)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedReport.ID, result.ID)
	assert.Equal(suite.T(), expectedReport.Title, result.Title)

	suite.mockRepo.AssertExpectations(suite.T())
}

// TestGetReportByID_NotFound prueba reporte no encontrado
func (suite *ReportServiceTestSuite) TestGetReportByID_NotFound() {
	// Configurar mock para devolver error
	suite.mockRepo.On("GetByID", 999).Return(nil, errors.New("report not found"))

	suite.mockLogger.On("Warn", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	result, err := suite.service.GetReportByID(999)

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "reporte no encontrado")

	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestGetReportsByUserID_Success prueba obtener reportes por usuario
func (suite *ReportServiceTestSuite) TestGetReportsByUserID_Success() {
	expectedReports := []*models.Report{
		{
			ID:          1,
			UserID:      1,
			Title:       "Reporte 1",
			Description: "Descripción 1",
			Status:      "pending",
		},
		{
			ID:          2,
			UserID:      1,
			Title:       "Reporte 2",
			Description: "Descripción 2",
			Status:      "resolved",
		},
	}

	// Configurar mock
	suite.mockRepo.On("GetByUserID", 1, 10, 0).Return(expectedReports, nil)

	// Ejecutar función
	result, err := suite.service.GetReportsByUserID(1, 10, 0)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), expectedReports[0].ID, result[0].ID)
	assert.Equal(suite.T(), expectedReports[1].ID, result[1].ID)

	suite.mockRepo.AssertExpectations(suite.T())
}

// TestGetReportsByUserID_EmptyResult prueba resultado vacío
func (suite *ReportServiceTestSuite) TestGetReportsByUserID_EmptyResult() {
	// Configurar mock para devolver slice vacío
	suite.mockRepo.On("GetByUserID", 1, 10, 0).Return([]*models.Report{}, nil)

	// Ejecutar función
	result, err := suite.service.GetReportsByUserID(1, 10, 0)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 0)

	suite.mockRepo.AssertExpectations(suite.T())
}

// TestUpdateReport_Success prueba actualización exitosa
func (suite *ReportServiceTestSuite) TestUpdateReport_Success() {
	reportToUpdate := &models.Report{
		ID:          1,
		UserID:      1,
		Title:       "Título actualizado",
		Description: "Descripción actualizada",
		Status:      "in_progress",
	}

	// Configurar mocks
	suite.mockRepo.On("Update", reportToUpdate).Return(nil)
	suite.mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	err := suite.service.UpdateReport(reportToUpdate)

	// Verificar resultado
	assert.NoError(suite.T(), err)

	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestUpdateReport_RepositoryError prueba error al actualizar
func (suite *ReportServiceTestSuite) TestUpdateReport_RepositoryError() {
	reportToUpdate := &models.Report{
		ID:     1,
		UserID: 1,
		Title:  "Test",
	}

	// Configurar mocks con error
	suite.mockRepo.On("Update", reportToUpdate).Return(errors.New("update failed"))
	suite.mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	err := suite.service.UpdateReport(reportToUpdate)

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "error al actualizar reporte")

	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestDeleteReport_Success prueba eliminación exitosa
func (suite *ReportServiceTestSuite) TestDeleteReport_Success() {
	// Configurar mocks
	suite.mockRepo.On("Delete", 1).Return(nil)
	suite.mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	err := suite.service.DeleteReport(1)

	// Verificar resultado
	assert.NoError(suite.T(), err)

	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestDeleteReport_RepositoryError prueba error al eliminar
func (suite *ReportServiceTestSuite) TestDeleteReport_RepositoryError() {
	// Configurar mocks con error
	suite.mockRepo.On("Delete", 999).Return(errors.New("report not found"))
	suite.mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	err := suite.service.DeleteReport(999)

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "error al eliminar reporte")

	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestGetReportsByStatus_Success prueba filtrar por estado
func (suite *ReportServiceTestSuite) TestGetReportsByStatus_Success() {
	expectedReports := []*models.Report{
		{
			ID:     1,
			UserID: 1,
			Status: "pending",
		},
		{
			ID:     2,
			UserID: 2,
			Status: "pending",
		},
	}

	// Configurar mock
	suite.mockRepo.On("GetByStatus", "pending", 10, 0).Return(expectedReports, nil)

	// Ejecutar función
	result, err := suite.service.GetReportsByStatus("pending", 10, 0)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "pending", result[0].Status)
	assert.Equal(suite.T(), "pending", result[1].Status)

	suite.mockRepo.AssertExpectations(suite.T())
}

// TestGetReportsByCategory_Success prueba filtrar por categoría
func (suite *ReportServiceTestSuite) TestGetReportsByCategory_Success() {
	expectedReports := []*models.Report{
		{
			ID:       1,
			UserID:   1,
			Category: "limpieza_publica",
		},
	}

	// Configurar mock
	suite.mockRepo.On("GetByCategory", "limpieza_publica", 10, 0).Return(expectedReports, nil)

	// Ejecutar función
	result, err := suite.service.GetReportsByCategory("limpieza_publica", 10, 0)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "limpieza_publica", result[0].Category)

	suite.mockRepo.AssertExpectations(suite.T())
}

// TestGetReportsByLocation_Success prueba filtrar por ubicación
func (suite *ReportServiceTestSuite) TestGetReportsByLocation_Success() {
	expectedReports := []*models.Report{
		{
			ID:     1,
			UserID: 1,
			Location: models.Location{
				Latitude:  -0.1807,
				Longitude: -78.4678,
			},
		},
	}

	// Configurar mock
	suite.mockRepo.On("GetByLocation", -0.1807, -78.4678, 1000.0, 10, 0).
		Return(expectedReports, nil)

	// Ejecutar función
	result, err := suite.service.GetReportsByLocation(-0.1807, -78.4678, 1000.0, 10, 0)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), -0.1807, result[0].Location.Latitude)
	assert.Equal(suite.T(), -78.4678, result[0].Location.Longitude)

	suite.mockRepo.AssertExpectations(suite.T())
}

// TestGetReportsCountByUserID_Success prueba contar reportes por usuario
func (suite *ReportServiceTestSuite) TestGetReportsCountByUserID_Success() {
	// Configurar mock
	suite.mockRepo.On("CountByUserID", 1).Return(5, nil)

	// Ejecutar función
	result, err := suite.service.GetReportsCountByUserID(1)

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, result)

	suite.mockRepo.AssertExpectations(suite.T())
}

// TestGetReportsCountByUserID_Error prueba error al contar
func (suite *ReportServiceTestSuite) TestGetReportsCountByUserID_Error() {
	// Configurar mock con error
	suite.mockRepo.On("CountByUserID", 999).Return(0, errors.New("database error"))

	suite.mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything)

	// Ejecutar función
	result, err := suite.service.GetReportsCountByUserID(999)

	// Verificar error
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), 0, result)
	assert.Contains(suite.T(), err.Error(), "error al obtener cantidad de reportes")

	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockLogger.AssertExpectations(suite.T())
}

// TestReportService ejecuta toda la suite
func TestReportService(t *testing.T) {
	suite.Run(t, new(ReportServiceTestSuite))
}

// Función auxiliar
func stringPtr(s string) *string {
	return &s
}
