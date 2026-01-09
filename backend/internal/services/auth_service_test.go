package services

import (
	"billing-note/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret", 24*time.Hour)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockRepo.On("FindByEmail", req.Email).Return(nil, errors.New("user not found"))
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	response, err := service.Register(req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.NotNil(t, response.User)
	assert.Equal(t, req.Email, response.User.Email)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret", 24*time.Hour)

	existingUser := &models.User{
		ID:    1,
		Email: "existing@example.com",
	}

	req := &RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockRepo.On("FindByEmail", req.Email).Return(existingUser, nil)

	response, err := service.Register(req)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "email already registered", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_CreateError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret", 24*time.Hour)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockRepo.On("FindByEmail", req.Email).Return(nil, errors.New("user not found"))
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(errors.New("database error"))

	response, err := service.Register(req)
	assert.Error(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret", 24*time.Hour)

	password := "password123"
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	user.SetPassword(password)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	mockRepo.On("FindByEmail", req.Email).Return(user, nil)

	response, err := service.Login(req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, user.Email, response.User.Email)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret", 24*time.Hour)

	req := &LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	mockRepo.On("FindByEmail", req.Email).Return(nil, errors.New("user not found"))

	response, err := service.Login(req)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid email or password", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret", 24*time.Hour)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
	}
	user.SetPassword("correctpassword")

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockRepo.On("FindByEmail", req.Email).Return(user, nil)

	response, err := service.Login(req)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, "invalid email or password", err.Error())

	mockRepo.AssertExpectations(t)
}
