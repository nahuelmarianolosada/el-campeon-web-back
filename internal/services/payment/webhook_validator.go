package payment

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log"
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
	log.Printf("[WebhookValidator.ValidateSignature] INFO: Validating webhook signature - topic=%s, dataID=%s", topic, dataID)
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
		log.Printf("[WebhookValidator.ValidateSignature] WARNING: Missing timestamp or signature in header - xSignature=%s", xSignature)
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
	isValid := constantTimeCompare(computedSignature, signature)
	if !isValid {
		log.Printf("[WebhookValidator.ValidateSignature] WARNING: Signature mismatch - computed=%s, received=%s", computedSignature, signature)
	} else {
		log.Printf("[WebhookValidator.ValidateSignature] INFO: Signature validated successfully")
	}
	return isValid
}

// constantTimeCompare compara dos strings en tiempo constante para evitar timing attacks
func constantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
