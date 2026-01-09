package handlers

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/internal/services"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionService is a mock implementation of TransactionService
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) CreateTransaction(userID uint, req *services.CreateTransactionRequest) (*models.Transaction, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransaction(id uint, userID uint) (*models.Transaction, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) ListTransactions(userID uint, filter repository.TransactionFilter) ([]models.Transaction, int64, error) {
	args := m.Called(userID, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionService) UpdateTransaction(id uint, userID uint, req *services.UpdateTransactionRequest) (*models.Transaction, error) {
	args := m.Called(id, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionService) DeleteTransaction(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockTransactionService) GetMonthlyStats(userID uint, year int, month int) (map[string]float64, error) {
	args := m.Called(userID, year, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockTransactionService) GetCategoryStats(userID uint, startDate, endDate time.Time, transactionType string) ([]map[string]interface{}, error) {
	args := m.Called(userID, startDate, endDate, transactionType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func setupTransactionTest() (*gin.Engine, *MockTransactionService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router.POST("/transactions", handler.Create)
	router.GET("/transactions/:id", handler.Get)
	router.GET("/transactions", handler.List)
	router.PUT("/transactions/:id", handler.Update)
	router.DELETE("/transactions/:id", handler.Delete)
	router.GET("/transactions/stats/monthly", handler.GetMonthlyStats)
	router.GET("/transactions/stats/category", handler.GetCategoryStats)

	return router, mockService
}

func createTestContext(userID uint) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID)
	return c
}

func TestTransactionHandler_Create_Success(t *testing.T) {
	router, mockService := setupTransactionTest()

	categoryID := uint(1)
	createReq := &services.CreateTransactionRequest{
		CategoryID:      &categoryID,
		Amount:          100.50,
		Type:            "expense",
		Description:     "Test transaction",
		TransactionDate: time.Now(),
		Source:          "manual",
	}

	mockTransaction := &models.Transaction{
		ID:              1,
		UserID:          1,
		CategoryID:      &categoryID,
		Amount:          100.50,
		Type:            "expense",
		Description:     "Test transaction",
		TransactionDate: time.Now(),
		Source:          "manual",
	}

	mockService.On("CreateTransaction", uint(1), mock.AnythingOfType("*services.CreateTransactionRequest")).
		Return(mockTransaction, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(mockService)
	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, 100.50, response.Amount)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Create_Failure_Unauthorized(t *testing.T) {
	router, _ := setupTransactionTest()

	createReq := map[string]interface{}{
		"amount": 100.50,
		"type":   "expense",
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTransactionHandler_Create_Failure_InvalidJSON(t *testing.T) {
	router, _ := setupTransactionTest()

	req, _ := http.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(new(MockTransactionService))
	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_Get_Success(t *testing.T) {
	router, mockService := setupTransactionTest()

	categoryID := uint(1)
	mockTransaction := &models.Transaction{
		ID:              1,
		UserID:          1,
		CategoryID:      &categoryID,
		Amount:          100.50,
		Type:            "expense",
		Description:     "Test transaction",
		TransactionDate: time.Now(),
	}

	mockService.On("GetTransaction", uint(1), uint(1)).Return(mockTransaction, nil)

	req, _ := http.NewRequest(http.MethodGet, "/transactions/1", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := NewTransactionHandler(mockService)
	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), response.ID)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Get_Failure_NotFound(t *testing.T) {
	router, mockService := setupTransactionTest()

	mockService.On("GetTransaction", uint(999), uint(1)).
		Return(nil, errors.New("transaction not found"))

	req, _ := http.NewRequest(http.MethodGet, "/transactions/999", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler := NewTransactionHandler(mockService)
	handler.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Get_Failure_InvalidID(t *testing.T) {
	router, _ := setupTransactionTest()

	req, _ := http.NewRequest(http.MethodGet, "/transactions/invalid", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler := NewTransactionHandler(new(MockTransactionService))
	handler.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_List_Success(t *testing.T) {
	router, mockService := setupTransactionTest()

	mockTransactions := []models.Transaction{
		{ID: 1, Amount: 100.50, Type: "expense"},
		{ID: 2, Amount: 200.00, Type: "income"},
	}

	mockService.On("ListTransactions", uint(1), mock.AnythingOfType("repository.TransactionFilter")).
		Return(mockTransactions, int64(2), nil)

	req, _ := http.NewRequest(http.MethodGet, "/transactions?page=1&page_size=10", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(mockService)
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["total"])

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_List_WithFilters(t *testing.T) {
	router, mockService := setupTransactionTest()

	mockTransactions := []models.Transaction{
		{ID: 1, Amount: 100.50, Type: "expense"},
	}

	mockService.On("ListTransactions", uint(1), mock.AnythingOfType("repository.TransactionFilter")).
		Return(mockTransactions, int64(1), nil)

	req, _ := http.NewRequest(http.MethodGet, "/transactions?type=expense&start_date=2024-01-01&end_date=2024-12-31&category_id=1", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(mockService)
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Update_Success(t *testing.T) {
	router, mockService := setupTransactionTest()

	updateReq := &services.UpdateTransactionRequest{
		Amount:      150.00,
		Type:        "income",
		Description: "Updated",
	}

	categoryID := uint(1)
	mockTransaction := &models.Transaction{
		ID:          1,
		UserID:      1,
		CategoryID:  &categoryID,
		Amount:      150.00,
		Type:        "income",
		Description: "Updated",
	}

	mockService.On("UpdateTransaction", uint(1), uint(1), mock.AnythingOfType("*services.UpdateTransactionRequest")).
		Return(mockTransaction, nil)

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/transactions/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := NewTransactionHandler(mockService)
	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 150.00, response.Amount)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Update_Failure_NotFound(t *testing.T) {
	router, mockService := setupTransactionTest()

	updateReq := &services.UpdateTransactionRequest{
		Amount: 150.00,
	}

	mockService.On("UpdateTransaction", uint(999), uint(1), mock.AnythingOfType("*services.UpdateTransactionRequest")).
		Return(nil, errors.New("transaction not found"))

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/transactions/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler := NewTransactionHandler(mockService)
	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Delete_Success(t *testing.T) {
	router, mockService := setupTransactionTest()

	mockService.On("DeleteTransaction", uint(1), uint(1)).Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/transactions/1", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler := NewTransactionHandler(mockService)
	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "transaction deleted successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Delete_Failure(t *testing.T) {
	router, mockService := setupTransactionTest()

	mockService.On("DeleteTransaction", uint(999), uint(1)).
		Return(errors.New("transaction not found"))

	req, _ := http.NewRequest(http.MethodDelete, "/transactions/999", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler := NewTransactionHandler(mockService)
	handler.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_GetMonthlyStats_Success(t *testing.T) {
	router, mockService := setupTransactionTest()

	mockStats := map[string]float64{
		"total_income":  1000.00,
		"total_expense": 500.00,
		"balance":       500.00,
	}

	mockService.On("GetMonthlyStats", uint(1), 2024, 1).Return(mockStats, nil)

	req, _ := http.NewRequest(http.MethodGet, "/transactions/stats/monthly?year=2024&month=1", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(mockService)
	handler.GetMonthlyStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]float64
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1000.00, response["total_income"])
	assert.Equal(t, 500.00, response["total_expense"])

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_GetMonthlyStats_DefaultValues(t *testing.T) {
	router, mockService := setupTransactionTest()

	now := time.Now()
	mockStats := map[string]float64{
		"total_income":  1000.00,
		"total_expense": 500.00,
	}

	mockService.On("GetMonthlyStats", uint(1), now.Year(), int(now.Month())).Return(mockStats, nil)

	req, _ := http.NewRequest(http.MethodGet, "/transactions/stats/monthly", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(mockService)
	handler.GetMonthlyStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_GetCategoryStats_Success(t *testing.T) {
	router, mockService := setupTransactionTest()

	mockStats := []map[string]interface{}{
		{"category": "Food", "total": 300.00, "count": 5},
		{"category": "Transport", "total": 200.00, "count": 3},
	}

	startDate, _ := time.Parse("2006-01-02", "2024-01-01")
	endDate, _ := time.Parse("2006-01-02", "2024-12-31")

	mockService.On("GetCategoryStats", uint(1), startDate, endDate, "expense").
		Return(mockStats, nil)

	req, _ := http.NewRequest(http.MethodGet, "/transactions/stats/category?start_date=2024-01-01&end_date=2024-12-31&type=expense", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(mockService)
	handler.GetCategoryStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response))

	mockService.AssertExpectations(t)
}

func TestTransactionHandler_GetCategoryStats_Failure_InvalidDate(t *testing.T) {
	router, _ := setupTransactionTest()

	req, _ := http.NewRequest(http.MethodGet, "/transactions/stats/category?start_date=invalid&end_date=2024-12-31&type=expense", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler := NewTransactionHandler(new(MockTransactionService))
	handler.GetCategoryStats(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
