package payment

import (
	"testing"
)

func TestWebhookValidatorValidSignature(t *testing.T) {
	tests := []struct {
		name          string
		publicKey     string
		xSignature    string
		dataID        string
		xRequestId    string
		secret        string
		expectedValid bool
	}{
		{
			name:          "Real world example from issue",
			publicKey:     "test_public_key",
			xSignature:    "ts=1779638450,v1=925fd4d121f45f782c7d5f9beb0f0697b938e15e3279170d74243e6b0e66238c",
			dataID:        "1327248690",
			xRequestId:    "f00fb98c-d3a5-4948-bf01-b8a29ae95c02",
			secret:        "dummy_secret_for_now", // I don't know the secret that produced that signature
			expectedValid: false,
		},
		{
			name:          "Invalid signature format",
			publicKey:     "test_public_key",
			xSignature:    "ts=1778367737,v1=invalidsignature",
			dataID:        "1346501131",
			xRequestId:    "req-123",
			secret:        "test_secret",
			expectedValid: false,
		},
		{
			name:          "Missing timestamp",
			publicKey:     "test_public_key",
			xSignature:    "v1=4d57042cf9734e2b92dc5c336f294a405f45bd0a6b3635af1b431ece26a51f59",
			dataID:        "1346501131",
			xRequestId:    "req-123",
			secret:        "test_secret",
			expectedValid: false,
		},
		{
			name:      "Valid signature (calculated)",
			publicKey: "test_public_key",
			// id:123;request-id:req-456;ts:1700000000;
			// secret: secret123
			// hmac-sha256("id:123;request-id:req-456;ts:1700000000;", "secret123")
			// Echo -n "id:123;request-id:req-456;ts:1700000000;" | openssl dgst -sha256 -hmac "secret123"
			// Result: 1b9178d79efa9ef302f996e30115bc5c14f3d7241ab18817456f2c49d01fb86c
			xSignature:    "ts=1700000000,v1=1b9178d79efa9ef302f996e30115bc5c14f3d7241ab18817456f2c49d01fb86c",
			dataID:        "123",
			xRequestId:    "req-456",
			secret:        "secret123",
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewWebhookValidator(tt.publicKey)
			isValid := validator.ValidateSignature(tt.xSignature, tt.dataID, tt.xRequestId, tt.secret)

			if isValid != tt.expectedValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectedValid, isValid)
			}
		})
	}
}

func TestWebhookValidatorConstantTimeCompare(t *testing.T) {
	// constantTimeCompare was replaced by hmac.Equal in the implementation
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
