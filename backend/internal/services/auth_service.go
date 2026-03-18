package services

import (
	"billing-note/internal/models"
	"billing-note/internal/repository"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"billing-note/pkg/utils"
	"time"
)

type AuthService interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *authService) Register(req *RegisterRequest) (*AuthResponse, error) {
	log := logger.ServiceLog("AuthService", "Register")

	log.WithField("email", req.Email).Debug("Checking if email is already registered")

	// Check if user already exists
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		log.WithField("email", req.Email).Warn("Registration failed: email already registered")
		return nil, errors.NewConflictError("Email address is already registered. Please use a different email or login to your existing account.")
	}

	// Create new user
	user := &models.User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: req.Password, // Will be hashed in BeforeCreate hook
	}

	log.WithField("email", req.Email).Debug("Creating new user")

	if err := s.userRepo.Create(user); err != nil {
		log.WithFields(logger.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Error("Failed to create user in database")
		return nil, errors.NewDBError("user creation", err)
	}

	log.WithFields(logger.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Debug("User created, generating JWT token")

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		log.WithFields(logger.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate JWT token")
		return nil, errors.NewInternalError("Failed to generate authentication token", err)
	}

	log.WithFields(logger.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User registered successfully")

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	log := logger.ServiceLog("AuthService", "Login")

	log.WithField("email", req.Email).Debug("Attempting user login")

	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		log.WithFields(logger.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Warn("Login failed: user not found")
		return nil, errors.NewInvalidCredentialsError()
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		log.WithFields(logger.Fields{
			"email":   req.Email,
			"user_id": user.ID,
		}).Warn("Login failed: invalid password")
		return nil, errors.NewInvalidCredentialsError()
	}

	log.WithFields(logger.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Debug("Password verified, generating JWT token")

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		log.WithFields(logger.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate JWT token")
		return nil, errors.NewInternalError("Failed to generate authentication token", err)
	}

	log.WithFields(logger.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User logged in successfully")

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}
