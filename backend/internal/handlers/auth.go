package handlers

import (
	"billing-note/internal/services"
	"billing-note/pkg/errors"
	"billing-note/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	log := logger.APILog("AuthHandler", "Register")
	requestID := c.GetString("request_id")

	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Invalid registration request: failed to parse JSON body")

		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"email":      req.Email,
		"name":       req.Name,
	}).Info("Processing user registration")

	response, err := h.authService.Register(&req)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"email":      req.Email,
			"error":      err.Error(),
		}).Warn("Registration failed")

		// Check if it's an AppError
		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			c.JSON(http.StatusBadRequest, errors.ErrorResponse{
				Error:   err.Error(),
				TraceID: requestID,
			})
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    response.User.ID,
		"email":      response.User.Email,
	}).Info("User registered successfully")

	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	log := logger.APILog("AuthHandler", "Login")
	requestID := c.GetString("request_id")

	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Invalid login request: failed to parse JSON body")

		appErr := errors.NewValidationError("Invalid request body: " + err.Error())
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"email":      req.Email,
	}).Info("Processing login attempt")

	response, err := h.authService.Login(&req)
	if err != nil {
		log.WithFields(logger.Fields{
			"request_id": requestID,
			"email":      req.Email,
			"error":      err.Error(),
		}).Warn("Login failed")

		if appErr := errors.GetAppError(err); appErr != nil {
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		} else {
			appErr := errors.NewInvalidCredentialsError()
			c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		}
		return
	}

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    response.User.ID,
		"email":      response.User.Email,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Me(c *gin.Context) {
	log := logger.APILog("AuthHandler", "Me")
	requestID := c.GetString("request_id")

	userID, exists := c.Get("user_id")
	if !exists {
		log.WithFields(logger.Fields{
			"request_id": requestID,
		}).Warn("Unauthorized access attempt to /me endpoint")

		appErr := errors.NewUnauthorizedError("User not authenticated")
		c.JSON(appErr.HTTPStatus, appErr.ToResponse(requestID))
		return
	}

	email, _ := c.Get("user_email")

	log.WithFields(logger.Fields{
		"request_id": requestID,
		"user_id":    userID,
	}).Debug("User profile retrieved")

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   email,
	})
}
