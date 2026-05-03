package payment

import (
	"context"

	orderMp "github.com/mercadopago/sdk-go/pkg/order"
)

// MercadopagoClient interface para abstraer el cliente de MercadoPago
type MercadopagoClient interface {
	CreatePayment(ctx context.Context, req orderMp.Request) (*orderMp.Response, error)
}

// DefaultMercadopagoClient implementa MercadopagoClient usando el SDK oficial
type DefaultMercadopagoClient struct {
	client orderMp.Client
}

func NewDefaultMercadopagoClient(client orderMp.Client) MercadopagoClient {
	return &DefaultMercadopagoClient{client: client}
}

func (c *DefaultMercadopagoClient) CreatePayment(ctx context.Context, req orderMp.Request) (*orderMp.Response, error) {
	return c.client.Create(ctx, req)
}

func (c *DefaultMercadopagoClient) CancelPayment(ctx context.Context, paymentID string) (*orderMp.Response, error) {
	return c.client.Cancel(ctx, paymentID)
}
