package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Payment struct {
	ID                      uint              `gorm:"primaryKey" json:"id"`
	TransactionID           string            `gorm:"type:varchar(255);uniqueIndex" json:"transaction_id"`
	OrderID                 uint              `gorm:"not null;index" json:"order_id"`
	Order                   *Order            `gorm:"foreignKey:OrderID" json:"-"`
	UserID                  uint              `gorm:"not null;index" json:"user_id"`
	User                    *User             `gorm:"foreignKey:UserID" json:"-"`
	Amount                  float64           `gorm:"not null" json:"amount"`
	Currency                string            `gorm:"default:'ARS'" json:"currency"`
	Status                  string            `gorm:"type:ENUM('PENDING','APPROVED','REJECTED','CANCELLED','REFUNDED');default:'PENDING'" json:"status"`
	PaymentMethod           string            `gorm:"type:ENUM('MP_SAVED','MP_INSTALLMENTS','MP_CARD','CASH');default:'MP_CARD'" json:"payment_method"`
	MercadopagoPreferenceID string            `json:"mercadopago_preference_id"`
	MercadopagoPaymentID    string            `json:"mercadopago_payment_id"`
	MercadopagoData         datatypes.JSONMap `gorm:"type:JSON" json:"mercadopago_data"`
	ApprovedAt              *time.Time        `json:"approved_at"`
	RejectedReason          string            `json:"rejected_reason"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
	DeletedAt               gorm.DeletedAt    `gorm:"index" json:"-"`
}

type CreatePaymentRequest struct {
	OrderID       uint    `json:"order_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=MP_SAVED MP_INSTALLMENTS MP_CARD CASH"`
}

type PaymentResponse struct {
	ID                      uint       `json:"id"`
	TransactionID           string     `json:"transaction_id"`
	OrderID                 uint       `json:"order_id"`
	UserID                  uint       `json:"user_id"`
	Amount                  float64    `json:"amount"`
	Currency                string     `json:"currency"`
	Status                  string     `json:"status"`
	PaymentMethod           string     `json:"payment_method"`
	MercadopagoPreferenceID string     `json:"mercadopago_preference_id"`
	ApprovedAt              *time.Time `json:"approved_at"`
	RejectedReason          string     `json:"rejected_reason"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// MercadopagoWebhookRequest estructura para recibir webhooks de MercadoPago
type MercadopagoWebhookRequest struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Action      string `json:"action"`
	APIVersion  string `json:"api_version"`
	DateCreated string `json:"date_created"`
	LiveMode    bool   `json:"live_mode"`
	UserID      string `json:"user_id"`
	Data        struct {
		ID string `json:"id"`
	} `json:"data"`
}

// MercadopagoPaymentDetailsResponse estructura para la respuesta de detalles del pago de MP
type MercadopagoPaymentDetailsResponse struct {
	ID                int64   `json:"id"`
	Status            string  `json:"status"` // approved, rejected, pending, cancelled, refunded, charged_back
	StatusDetail      string  `json:"status_detail"`
	TransactionAmount float64 `json:"transaction_amount"`
	Currency          string  `json:"currency_id"`
	PaymentMethod     struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"payment_method"`
	DateCreated       string `json:"date_created"`
	DateLastModified  string `json:"date_last_modified"`
	ExternalReference string `json:"external_reference"`
}
