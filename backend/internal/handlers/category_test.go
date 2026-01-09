package handlers

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCategoryRepository is a mock implementation of CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetAll() ([]models.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByType(categoryType string) ([]models.Category, error) {
	args := m.Called(categoryType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByID(id uint) (*models.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Create(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Update(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *MockCategoryRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupCategoryTest() (*gin.Engine, *MockCategoryRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockRepo := new(MockCategoryRepository)
	handler := NewCategoryHandler(mockRepo)

	router.GET("/categories", handler.GetAll)
	router.GET("/categories/:type", handler.GetByType)

	return router, mockRepo
}

func TestCategoryHandler_GetAll_Success(t *testing.T) {
	router, mockRepo := setupCategoryTest()

	mockCategories := []models.Category{
		{ID: 1, Name: "Food", Type: "expense", Icon: "🍔"},
		{ID: 2, Name: "Salary", Type: "income", Icon: "💰"},
		{ID: 3, Name: "Transport", Type: "expense", Icon: "🚗"},
	}

	mockRepo.On("GetAll").Return(mockCategories, nil)

	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Category
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(response))
	assert.Equal(t, "Food", response[0].Name)
	assert.Equal(t, "Salary", response[1].Name)

	mockRepo.AssertExpectations(t)
}

func TestCategoryHandler_GetAll_Failure_DatabaseError(t *testing.T) {
	router, mockRepo := setupCategoryTest()

	mockRepo.On("GetAll").Return(nil, errors.New("database connection error"))

	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database connection error", response["error"])

	mockRepo.AssertExpectations(t)
}

func TestCategoryHandler_GetAll_EmptyList(t *testing.T) {
	router, mockRepo := setupCategoryTest()

	mockCategories := []models.Category{}

	mockRepo.On("GetAll").Return(mockCategories, nil)

	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Category
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(response))

	mockRepo.AssertExpectations(t)
}

func TestCategoryHandler_GetByType_Success_Income(t *testing.T) {
	router, mockRepo := setupCategoryTest()

	mockCategories := []models.Category{
		{ID: 2, Name: "Salary", Type: "income", Icon: "💰"},
		{ID: 4, Name: "Bonus", Type: "income", Icon: "🎁"},
	}

	mockRepo.On("GetByType", "income").Return(mockCategories, nil)

	req, _ := http.NewRequest(http.MethodGet, "/categories/income", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Category
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "income", response[0].Type)
	assert.Equal(t, "income", response[1].Type)

	mockRepo.AssertExpectations(t)
}

func TestCategoryHandler_GetByType_Success_Expense(t *testing.T) {
	router, mockRepo := setupCategoryTest()

	mockCategories := []models.Category{
		{ID: 1, Name: "Food", Type: "expense", Icon: "🍔"},
		{ID: 3, Name: "Transport", Type: "expense", Icon: "🚗"},
	}

	mockRepo.On("GetByType", "expense").Return(mockCategories, nil)

	req, _ := http.NewRequest(http.MethodGet, "/categories/expense", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Category
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "expense", response[0].Type)

	mockRepo.AssertExpectations(t)
}

func TestCategoryHandler_GetByType_Failure_InvalidType(t *testing.T) {
	router, _ := setupCategoryTest()

	req, _ := http.NewRequest(http.MethodGet, "/categories/invalid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "type must be 'income' or 'expense'", response["error"])
}

func TestCategoryHandler_GetByType_Failure_DatabaseError(t *testing.T) {
	router, mockRepo := setupCategoryTest()

	mockRepo.On("GetByType", "income").Return(nil, errors.New("database error"))

	req, _ := http.NewRequest(http.MethodGet, "/categories/income", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database error", response["error"])

	mockRepo.AssertExpectations(t)
}

func TestCategoryHandler_GetByType_EmptyList(t *testing.T) {
	router, mockRepo := setupCategoryTest()

	mockCategories := []models.Category{}

	mockRepo.On("GetByType", "income").Return(mockCategories, nil)

	req, _ := http.NewRequest(http.MethodGet, "/categories/income", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Category
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(response))

	mockRepo.AssertExpectations(t)
}
