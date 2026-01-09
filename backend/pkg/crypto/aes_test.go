package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAESCrypto(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid 32-byte key",
			key:     "12345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:    "short key - will be padded",
			key:     "short",
			wantErr: false,
		},
		{
			name:    "long key - will be truncated",
			key:     "1234567890123456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "empty key - will be padded",
			key:     "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crypto, err := NewAESCrypto(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, crypto)
			}
		})
	}
}

func TestAESCrypto_EncryptDecrypt(t *testing.T) {
	crypto, err := NewAESCrypto("test-encryption-key-32-bytes!!!")
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "hello world",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "password-like string",
			plaintext: "F128217250",
		},
		{
			name:      "unicode text",
			plaintext: "密碼測試123",
		},
		{
			name:      "long text",
			plaintext: "This is a very long text that should be encrypted properly even though it exceeds the block size of AES",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := crypto.Encrypt(tt.plaintext)
			assert.NoError(t, err)
			assert.NotEmpty(t, ciphertext)
			assert.NotEqual(t, tt.plaintext, ciphertext)

			// Decrypt
			decrypted, err := crypto.Decrypt(ciphertext)
			assert.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestAESCrypto_DecryptInvalid(t *testing.T) {
	crypto, err := NewAESCrypto("test-key")
	require.NoError(t, err)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!!!",
		},
		{
			name:       "too short",
			ciphertext: "YWJj", // "abc" in base64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := crypto.Decrypt(tt.ciphertext)
			assert.Error(t, err)
		})
	}
}

func TestAESCrypto_DifferentCiphertexts(t *testing.T) {
	crypto, err := NewAESCrypto("test-key")
	require.NoError(t, err)

	plaintext := "same text"

	// Encrypt same text twice
	cipher1, err := crypto.Encrypt(plaintext)
	require.NoError(t, err)

	cipher2, err := crypto.Encrypt(plaintext)
	require.NoError(t, err)

	// Ciphertexts should be different due to random nonce
	assert.NotEqual(t, cipher1, cipher2)

	// But both should decrypt to the same plaintext
	decrypted1, err := crypto.Decrypt(cipher1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := crypto.Decrypt(cipher2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}
