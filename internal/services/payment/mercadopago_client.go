package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	preferenceMp "github.com/mercadopago/sdk-go/pkg/preference"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
)

// MercadopagoClient interface para abstraer el cliente de MercadoPago
type MercadopagoClient interface {
	CreatePreference(ctx context.Context, req preferenceMp.Request) (*preferenceMp.Response, error)
	GetPaymentDetails(ctx context.Context, paymentID string) (*models.MercadopagoPaymentDetailsResponse, error)
}

// DefaultMercadopagoClient implementa MercadopagoClient usando el SDK oficial
type DefaultMercadopagoClient struct {
	preferenceClient preferenceMp.Client
	accessToken      string
}

func NewDefaultMercadopagoClient(client preferenceMp.Client, accessToken string) MercadopagoClient {
	return &DefaultMercadopagoClient{
		preferenceClient: client,
		accessToken:      accessToken,
	}
}

func (c *DefaultMercadopagoClient) CreatePreference(ctx context.Context, req preferenceMp.Request) (*preferenceMp.Response, error) {
	log.Printf("[DefaultMercadopagoClient.CreatePreference] INFO: Creating preference")
	resp, err := c.preferenceClient.Create(ctx, req)
	if err != nil {
		log.Printf("[DefaultMercadopagoClient.CreatePreference] ERROR: Failed to create preference: %v", err)
		return nil, err
	}
	log.Printf("[DefaultMercadopagoClient.CreatePreference] INFO: Preference created successfully - ID=%s", resp.ID)
	return resp, nil
}

func (c *DefaultMercadopagoClient) GetPaymentDetails(ctx context.Context, paymentID string) (*models.MercadopagoPaymentDetailsResponse, error) {
	log.Printf("[DefaultMercadopagoClient.GetPaymentDetails] INFO: Retrieving payment details - mpPaymentID=%s", paymentID)
	// URL de la API de MercadoPago para obtener detalles del pago
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", paymentID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("[DefaultMercadopagoClient.GetPaymentDetails] ERROR: Failed to create request: %v", err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Agregar headers de autenticación
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("X-Idempotency-Key", fmt.Sprintf("webhook-payment-%s", paymentID))

	// Ejecutar request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[DefaultMercadopagoClient.GetPaymentDetails] ERROR: Failed to execute request: %v", err)
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[DefaultMercadopagoClient.GetPaymentDetails] ERROR: Failed to read response body: %v", err)
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Verificar código de estado
	if resp.StatusCode != http.StatusOK {
		log.Printf("[DefaultMercadopagoClient.GetPaymentDetails] ERROR: MP API returned non-OK status - status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("mercadopago api returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parsear JSON
	var paymentDetails models.MercadopagoPaymentDetailsResponse
	if err := json.Unmarshal(body, &paymentDetails); err != nil {
		log.Printf("[DefaultMercadopagoClient.GetPaymentDetails] ERROR: Failed to unmarshal response: %v", err)
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("[DefaultMercadopagoClient.GetPaymentDetails] INFO: Payment details retrieved successfully - mpPaymentID=%d, status=%s", paymentDetails.ID, paymentDetails.Status)
	return &paymentDetails, nil
}
