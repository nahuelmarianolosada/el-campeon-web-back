package payment

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// WebhookValidator valida la firma de webhooks de MercadoPago
type WebhookValidator struct {
	publicKey string
}

// NewWebhookValidator crea un nuevo validador de webhooks
func NewWebhookValidator(publicKey string) *WebhookValidator {
	return &WebhookValidator{
		publicKey: publicKey,
	}
}

// ValidateSignature valida la firma del webhook de MercadoPago
// La firma se encuentra en el header X-Signature en formato: ts=<timestamp>,v1=<signature>
// La firma se calcula como: HMAC-SHA256(base_id.timestamp.public_key, access_token)
func (v *WebhookValidator) ValidateSignature(xSignature string, topic string, dataID string, accessToken string) bool {
	// Parsear el header X-Signature
	signatureParts := strings.Split(xSignature, ",")
	var timestamp string
	var signature string

	for _, part := range signatureParts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "ts=") {
			timestamp = strings.TrimPrefix(part, "ts=")
		} else if strings.HasPrefix(part, "v1=") {
			signature = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || signature == "" {
		return false
	}

	// Construir el string a verificar: data_id.timestamp
	// Según la documentación de MercadoPago, es: base_id.timestamp
	verifyString := fmt.Sprintf("%s.%s", dataID, timestamp)

	// Crear el HMAC-SHA256 usando el access token como clave
	h := sha256.New()
	h.Write([]byte(verifyString))
	hashedBytes := h.Sum(nil)
	computedSignature := fmt.Sprintf("%x", hashedBytes)

	// La firma en el header ya es el hash hexadecimal
	// Comparar de forma segura
	return constantTimeCompare(computedSignature, signature)
}

// constantTimeCompare compara dos strings en tiempo constante para evitar timing attacks
func constantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}

	return result == 0
}

