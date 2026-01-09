package integration

import (
	"billing-note/pkg/config"
	"billing-note/pkg/database"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Set test environment variables
	os.Setenv("DB_NAME", "billing_note_test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_SSLMODE", "disable")

	cfg, err := config.Load()
	if err != nil {
		t.Skip("Skipping integration test: unable to load config")
		return nil
	}

	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		t.Skip("Skipping integration test: unable to connect to database")
		return nil
	}

	return db
}

func TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	assert.NoError(t, err)

	// Test connection
	err = sqlDB.Ping()
	assert.NoError(t, err)

	// Close connection
	err = sqlDB.Close()
	assert.NoError(t, err)
}

func TestDatabaseMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	if db == nil {
		return
	}

	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Test that tables exist
	assert.True(t, db.Migrator().HasTable("users"))
	assert.True(t, db.Migrator().HasTable("categories"))
	assert.True(t, db.Migrator().HasTable("transactions"))
}

func TestDatabaseTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	if db == nil {
		return
	}

	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Test transaction rollback
	tx := db.Begin()
	assert.NotNil(t, tx)

	// Do something in transaction
	err := tx.Exec("SELECT 1").Error
	assert.NoError(t, err)

	// Rollback
	err = tx.Rollback().Error
	assert.NoError(t, err)

	// Test transaction commit
	tx = db.Begin()
	assert.NotNil(t, tx)

	// Do something in transaction
	err = tx.Exec("SELECT 1").Error
	assert.NoError(t, err)

	// Commit
	err = tx.Commit().Error
	assert.NoError(t, err)
}

func TestDatabaseConnectionPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	if db == nil {
		return
	}

	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	sqlDB, err := db.DB()
	assert.NoError(t, err)

	// Get connection stats
	stats := sqlDB.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1)
}

func TestDatabaseConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	if db == nil {
		return
	}

	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Test concurrent queries
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			var result int
			err := db.Raw("SELECT 1").Scan(&result).Error
			assert.NoError(t, err)
			assert.Equal(t, 1, result)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDatabaseErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test connection with invalid credentials
	os.Setenv("DB_PASSWORD", "invalid_password")
	cfg, _ := config.Load()

	_, err := database.Connect(cfg.Database.DSN())
	assert.Error(t, err)
}
