package integration

import (
	"billing-note/internal/handlers"
	"billing-note/internal/middleware"
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/internal/services"
	"billing-note/pkg/config"
	"billing-note/pkg/database"
	"billing-note/pkg/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTransactionTestServer(t *testing.T) (*gin.Engine, *gorm.DB, string, uint, func()) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)

	// Setup test database
	os.Setenv("DB_NAME", "billing_note_test")
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Skipping integration test: unable to load config")
		return nil, nil, "", 0, nil
	}

	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		t.Skip("Skipping integration test: unable to connect to database")
		return nil, nil, "", 0, nil
	}

	// Create test user
	testUser := &models.User{
		Email:        fmt.Sprintf("transactiontest_%d@test.com", time.Now().Unix()),
		Name:         "Transaction Test User",
		PasswordHash: "password123",
	}
	db.Create(testUser)

	// Generate JWT token for test user
	token, _ := utils.GenerateToken(testUser.ID, testUser.Email, cfg.JWT.Secret, cfg.JWT.Expiry)

	// Setup repositories and services
	transactionRepo := repository.NewTransactionRepository(db)
	transactionService := services.NewTransactionService(transactionRepo)
	categoryRepo := repository.NewCategoryRepository(db)

	// Setup handlers
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	// Setup router
	router := gin.New()
	authMiddleware := middleware.AuthMiddleware(cfg.JWT.Secret)

	api := router.Group("/api")
	{
		api.POST("/transactions", authMiddleware, transactionHandler.Create)
		api.GET("/transactions/:id", authMiddleware, transactionHandler.Get)
		api.GET("/transactions", authMiddleware, transactionHandler.List)
		api.PUT("/transactions/:id", authMiddleware, transactionHandler.Update)
		api.DELETE("/transactions/:id", authMiddleware, transactionHandler.Delete)
		api.GET("/transactions/stats/monthly", authMiddleware, transactionHandler.GetMonthlyStats)
		api.GET("/transactions/stats/category", authMiddleware, transactionHandler.GetCategoryStats)
		api.GET("/categories", categoryHandler.GetAll)
		api.GET("/categories/:type", categoryHandler.GetByType)
	}

	// Cleanup function
	cleanup := func() {
		db.Exec("DELETE FROM transactions WHERE user_id = ?", testUser.ID)
		db.Exec("DELETE FROM users WHERE id = ?", testUser.ID)
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return router, db, token, testUser.ID, cleanup
}

func TestTransactionAPI_CreateFlow(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create a category first
	category := &models.Category{
		Name: "Test Category",
		Type: "expense",
		Icon: "🧪",
	}
	db.Create(category)

	// Create transaction
	createReq := map[string]interface{}{
		"category_id":      category.ID,
		"amount":           150.50,
		"type":             "expense",
		"description":      "Test transaction",
		"transaction_date": time.Now().Format(time.RFC3339),
		"source":           "manual",
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 150.50, response.Amount)
	assert.Equal(t, "expense", response.Type)

	// Verify transaction was created in database
	var transaction models.Transaction
	err = db.Where("user_id = ? AND amount = ?", userID, 150.50).First(&transaction).Error
	assert.NoError(t, err)
}

func TestTransactionAPI_GetFlow(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create a transaction
	transaction := &models.Transaction{
		UserID:          userID,
		Amount:          200.00,
		Type:            "income",
		Description:     "Get test",
		TransactionDate: time.Now(),
		Source:          "manual",
	}
	db.Create(transaction)

	// Get transaction
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/transactions/%d", transaction.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, transaction.ID, response.ID)
	assert.Equal(t, 200.00, response.Amount)
}

func TestTransactionAPI_ListFlow(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create multiple transactions
	for i := 0; i < 5; i++ {
		transaction := &models.Transaction{
			UserID:          userID,
			Amount:          float64(100 + i*10),
			Type:            "expense",
			Description:     fmt.Sprintf("Test transaction %d", i),
			TransactionDate: time.Now(),
			Source:          "manual",
		}
		db.Create(transaction)
	}

	// List transactions
	req, _ := http.NewRequest(http.MethodGet, "/api/transactions?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, int(response["total"].(float64)), 5)

	data := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(data), 5)
}

func TestTransactionAPI_ListWithFilters(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create transactions with different types
	db.Create(&models.Transaction{
		UserID:          userID,
		Amount:          100.00,
		Type:            "income",
		TransactionDate: time.Now(),
		Source:          "manual",
	})
	db.Create(&models.Transaction{
		UserID:          userID,
		Amount:          200.00,
		Type:            "expense",
		TransactionDate: time.Now(),
		Source:          "manual",
	})

	// Filter by type=income
	req, _ := http.NewRequest(http.MethodGet, "/api/transactions?type=income", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].([]interface{})
	for _, item := range data {
		trans := item.(map[string]interface{})
		assert.Equal(t, "income", trans["type"])
	}
}

func TestTransactionAPI_UpdateFlow(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create a transaction
	transaction := &models.Transaction{
		UserID:          userID,
		Amount:          100.00,
		Type:            "expense",
		Description:     "Original",
		TransactionDate: time.Now(),
		Source:          "manual",
	}
	db.Create(transaction)

	// Update transaction
	updateReq := map[string]interface{}{
		"amount":      250.00,
		"description": "Updated",
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/transactions/%d", transaction.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 250.00, response.Amount)
	assert.Equal(t, "Updated", response.Description)

	// Verify in database
	var updated models.Transaction
	db.First(&updated, transaction.ID)
	assert.Equal(t, 250.00, updated.Amount)
	assert.Equal(t, "Updated", updated.Description)
}

func TestTransactionAPI_DeleteFlow(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create a transaction
	transaction := &models.Transaction{
		UserID:          userID,
		Amount:          100.00,
		Type:            "expense",
		TransactionDate: time.Now(),
		Source:          "manual",
	}
	db.Create(transaction)

	// Delete transaction
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/transactions/%d", transaction.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deletion in database
	var deleted models.Transaction
	err := db.First(&deleted, transaction.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestTransactionAPI_GetMonthlyStats(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create transactions for current month
	now := time.Now()
	db.Create(&models.Transaction{
		UserID:          userID,
		Amount:          500.00,
		Type:            "income",
		TransactionDate: now,
		Source:          "manual",
	})
	db.Create(&models.Transaction{
		UserID:          userID,
		Amount:          300.00,
		Type:            "expense",
		TransactionDate: now,
		Source:          "manual",
	})

	// Get monthly stats
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/transactions/stats/monthly?year=%d&month=%d", now.Year(), int(now.Month())),
		nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]float64
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "total_income")
	assert.Contains(t, response, "total_expense")
}

func TestTransactionAPI_GetCategoryStats(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create category
	category := &models.Category{
		Name: "Food",
		Type: "expense",
		Icon: "🍔",
	}
	db.Create(category)

	// Create transactions
	db.Create(&models.Transaction{
		UserID:          userID,
		CategoryID:      &category.ID,
		Amount:          100.00,
		Type:            "expense",
		TransactionDate: time.Now(),
		Source:          "manual",
	})

	// Get category stats
	startDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/transactions/stats/category?start_date=%s&end_date=%s&type=expense", startDate, endDate),
		nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTransactionAPI_UnauthorizedAccess(t *testing.T) {
	router, _, _, _, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Try to access without token
	req, _ := http.NewRequest(http.MethodGet, "/api/transactions", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTransactionAPI_CrossUserAccess(t *testing.T) {
	router, db, token, userID, cleanup := setupTransactionTestServer(t)
	if router == nil {
		return
	}
	defer cleanup()

	// Create another user and their transaction
	anotherUser := &models.User{
		Email:        "another@test.com",
		Name:         "Another User",
		PasswordHash: "password",
	}
	db.Create(anotherUser)

	anotherTransaction := &models.Transaction{
		UserID:          anotherUser.ID,
		Amount:          500.00,
		Type:            "income",
		TransactionDate: time.Now(),
		Source:          "manual",
	}
	db.Create(anotherTransaction)

	// Try to access another user's transaction
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/transactions/%d", anotherTransaction.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return error (not found or unauthorized)
	assert.NotEqual(t, http.StatusOK, w.Code)

	// Cleanup
	db.Delete(anotherTransaction)
	db.Delete(anotherUser)
}
