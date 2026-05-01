package payment

import (
	"context"

	"github.com/mercadopago/sdk-go/pkg/preference"
)

// MercadopagoClient interface para abstraer el cliente de MercadoPago
type MercadopagoClient interface {
	CreatePreference(ctx context.Context, req preference.Request) (*preference.Response, error)
}

// DefaultMercadopagoClient implementa MercadopagoClient usando el SDK oficial
type DefaultMercadopagoClient struct {
	client preference.Client
}

func NewDefaultMercadopagoClient(client preference.Client) MercadopagoClient {
	return &DefaultMercadopagoClient{client: client}
}

func (c *DefaultMercadopagoClient) CreatePreference(ctx context.Context, req preference.Request) (*preference.Response, error) {
	return c.client.Create(ctx, req)
}


