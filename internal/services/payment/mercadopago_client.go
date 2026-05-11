package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	return c.preferenceClient.Create(ctx, req)
}

func (c *DefaultMercadopagoClient) GetPaymentDetails(ctx context.Context, paymentID string) (*models.MercadopagoPaymentDetailsResponse, error) {
	// URL de la API de MercadoPago para obtener detalles del pago
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", paymentID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Agregar headers de autenticación
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("X-Idempotency-Key", fmt.Sprintf("webhook-payment-%s", paymentID))

	// Ejecutar request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Verificar código de estado
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mercadopago api returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parsear JSON
	var paymentDetails models.MercadopagoPaymentDetailsResponse
	if err := json.Unmarshal(body, &paymentDetails); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &paymentDetails, nil
}

