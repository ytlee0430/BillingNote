package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUser_SetPassword(t *testing.T) {
	user := &User{}
	password := "testpassword123"

	err := user.SetPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, password, user.PasswordHash)

	// Verify the password hash is valid bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	assert.NoError(t, err)
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "testpassword123"

	err := user.SetPassword(password)
	assert.NoError(t, err)

	// Test correct password
	assert.True(t, user.CheckPassword(password))

	// Test incorrect password
	assert.False(t, user.CheckPassword("wrongpassword"))
}

func TestUser_CheckPassword_Empty(t *testing.T) {
	user := &User{
		PasswordHash: "",
	}

	assert.False(t, user.CheckPassword("anypassword"))
}

func TestUser_SetPassword_EmptyPassword(t *testing.T) {
	user := &User{}
	err := user.SetPassword("")
	assert.NoError(t, err)
}
