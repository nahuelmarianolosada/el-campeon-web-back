package payment

import (
	"testing"
)

func TestWebhookValidatorValidSignature(t *testing.T) {
	tests := []struct {
		name            string
		publicKey       string
		xSignature      string
		dataID          string
		accessToken     string
		expectedValid   bool
	}{
		{
			name:          "Invalid signature format",
			publicKey:     "test_public_key",
			xSignature:    "ts=1778367737,v1=invalidsignature",
			dataID:        "1346501131",
			accessToken:   "test_access_token",
			expectedValid: false, // Será false porque no es la firma real
		},
		{
			name:          "Missing timestamp",
			publicKey:     "test_public_key",
			xSignature:    "v1=4d57042cf9734e2b92dc5c336f294a405f45bd0a6b3635af1b431ece26a51f59",
			dataID:        "1346501131",
			accessToken:   "test_access_token",
			expectedValid: false,
		},
		{
			name:          "Missing signature",
			publicKey:     "test_public_key",
			xSignature:    "ts=1778367737",
			dataID:        "1346501131",
			accessToken:   "test_access_token",
			expectedValid: false,
		},
		{
			name:          "Empty signature header",
			publicKey:     "test_public_key",
			xSignature:    "",
			dataID:        "1346501131",
			accessToken:   "test_access_token",
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewWebhookValidator(tt.publicKey)
			isValid := validator.ValidateSignature(tt.xSignature, "payment", tt.dataID, tt.accessToken)

			if isValid != tt.expectedValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectedValid, isValid)
			}
		})
	}
}

func TestWebhookValidatorConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "Equal strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "Different strings same length",
			a:        "test",
			b:        "abcd",
			expected: false,
		},
		{
			name:     "Different lengths",
			a:        "test",
			b:        "testing",
			expected: false,
		},
		{
			name:     "Empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constantTimeCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestWebhookSignatureGeneration genera una firma de prueba válida
// Esto es solo con fines de demostración
func TestWebhookSignatureGeneration(t *testing.T) {
	// Estos son valores de ejemplo, en una prueba real usarías valores reales
	dataID := "1346501131"
	timestamp := "1778367737"
	accessToken := "test_access_token"

	validator := NewWebhookValidator("test_public_key")

	// Construir el string que se firmaría
	// verifyString := dataID + "." + timestamp
	// En un test real, compararíamos la firma generada con la esperada

	// Por ahora, simplemente verificamos que el código no paniquea
	isValid := validator.ValidateSignature(
		"ts="+timestamp+",v1=dummysignature",
		"payment",
		dataID,
		accessToken,
	)

	// isValid debería ser false porque la firma es dummy
	if isValid {
		t.Errorf("Expected false for dummy signature, got true")
	}
}


