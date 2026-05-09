package payment

import (
	"context"

	preferenceMp "github.com/mercadopago/sdk-go/pkg/preference"
)

// MercadopagoClient interface para abstraer el cliente de MercadoPago
type MercadopagoClient interface {
	CreatePreference(ctx context.Context, req preferenceMp.Request) (*preferenceMp.Response, error)
}

// DefaultMercadopagoClient implementa MercadopagoClient usando el SDK oficial
type DefaultMercadopagoClient struct {
	client preferenceMp.Client
}

func NewDefaultMercadopagoClient(client preferenceMp.Client) MercadopagoClient {
	return &DefaultMercadopagoClient{client: client}
}

func (c *DefaultMercadopagoClient) CreatePreference(ctx context.Context, req preferenceMp.Request) (*preferenceMp.Response, error) {
	return c.client.Create(ctx, req)
}

