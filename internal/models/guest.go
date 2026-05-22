package models

import "time"

type GuestSession struct {
	ID                       uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email                    string     `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	VerificationCodeHash     string     `gorm:"type:varchar(255)" json:"-"`
	VerificationCodeSentAt   *time.Time `json:"verification_code_sent_at,omitempty"`
	VerificationCodeAttempts int        `gorm:"default:0" json:"-"`
	IsVerified               bool       `gorm:"default:false" json:"is_verified"`
	VerifiedAt               *time.Time `json:"verified_at,omitempty"`
	GuestTokenHash           string     `gorm:"type:varchar(255)" json:"-"`
	SessionIPAddress         string     `gorm:"type:varchar(45)" json:"-"`
	UserID                   *uint      `gorm:"uniqueIndex;index" json:"user_id,omitempty"`
	AttemptsFromIP           int        `gorm:"default:0" json:"-"`
	LastAttemptAt            *time.Time `json:"-"`
	ExpiresAt                time.Time  `json:"expires_at"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}
type ConfirmEmailRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verification_code" binding:"required,len=6"`
}
type GuestSessionResponse struct {
	GuestToken string    `json:"guest_token"`
	Email      string    `json:"email"`
	ExpiresAt  time.Time `json:"expires_at"`
}
type GuestCartItem struct {
	SKU      string  `json:"sku" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price" binding:"required,gt=0"`
}
type CreateGuestOrderRequest struct {
	GuestName       string                 `json:"guest_name" binding:"required"`
	GuestEmail      string                 `json:"guest_email" binding:"required,email"`
	UserID          uint                   `json:"user_id,omitempty"`
	Items           []GuestCartItem        `json:"items" binding:"required,min=1"`
	ShippingAddress map[string]interface{} `json:"shipping_address" binding:"required"`
	DeliveryMethod  string                 `json:"delivery_method" binding:"required,oneof=shipping pickup-libreria pickup-jugueteria"`
	Notes           string                 `json:"notes"`
}
type CreateGuestPaymentRequest struct {
	OrderID       uint    `json:"order_id" binding:"required"`
	Email         string  `json:"email" binding:"required,email"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=MP_CARD MP_INSTALLMENTS CASH"`
}
