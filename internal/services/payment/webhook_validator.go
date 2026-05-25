package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
// La firma se calcula como: HMAC-SHA256(manifest, secret)
// manifest = id:[data.id];request-id:[x-request-id];ts:[ts];
func (v *WebhookValidator) ValidateSignature(
	xSignature string,
	dataID string,
	xRequestID string,
	secret string,
) bool {

	log.Printf(
		"[WebhookValidator.ValidateSignature] INFO: validating webhook signature - dataID=%s requestID=%s",
		dataID,
		xRequestID,
	)

	// Parse X-Signature
	var ts string
	var receivedSignatureHex string

	parts := strings.Split(xSignature, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, "ts=") {
			ts = strings.TrimPrefix(part, "ts=")
		}

		if strings.HasPrefix(part, "v1=") {
			receivedSignatureHex = strings.TrimPrefix(part, "v1=")
		}
	}

	if ts == "" || receivedSignatureHex == "" {
		log.Printf(
			"[WebhookValidator.ValidateSignature] ERROR: invalid X-Signature header: %s",
			xSignature,
		)
		return false
	}

	// IMPORTANT:
	// MercadoPago expects EXACT format:
	// id:[data.id];request-id:[x-request-id];ts:[ts];
	manifest := fmt.Sprintf(
		"id:%s;request-id:%s;ts:%s;",
		dataID,
		xRequestID,
		ts,
	)

	log.Printf(
		"[WebhookValidator.ValidateSignature] DEBUG: manifest=%s",
		manifest,
	)

	// Generate HMAC SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))

	expectedSignatureBytes := mac.Sum(nil)

	// Decode MercadoPago signature from hex
	receivedSignatureBytes, err := hex.DecodeString(receivedSignatureHex)
	if err != nil {
		log.Printf(
			"[WebhookValidator.ValidateSignature] ERROR: invalid signature hex: %v",
			err,
		)
		return false
	}

	// Constant-time comparison
	isValid := hmac.Equal(expectedSignatureBytes, receivedSignatureBytes)

	if !isValid {
		log.Printf(
			"[WebhookValidator.ValidateSignature] WARNING: signature mismatch expected=%s received=%s",
			hex.EncodeToString(expectedSignatureBytes),
			receivedSignatureHex,
		)
		return false
	}

	log.Printf(
		"[WebhookValidator.ValidateSignature] INFO: signature validated successfully",
	)

	return true
}
