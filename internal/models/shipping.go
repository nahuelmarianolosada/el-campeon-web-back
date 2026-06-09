package models

import (
	"time"

	"gorm.io/gorm"
)

type ZoneKind string

const (
	ZoneKindProvincial    ZoneKind = "provincial"
	ZoneKindRegional      ZoneKind = "regional"
	ZoneKindDepartamental ZoneKind = "departamental"
	ZoneKindBarrial       ZoneKind = "barrial"
)

type DeliveryZone struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(120);not null" json:"name"`
	Kind         string         `gorm:"type:ENUM('provincial','regional','departamental','barrial');not null" json:"kind"`
	ParentZoneID *uint          `gorm:"index" json:"parent_zone_id,omitempty"`
	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DeliveryZone) TableName() string { return "delivery_zones" }

type CreateZoneRequest struct {
	Name         string `json:"name" binding:"required,max=120"`
	Kind         string `json:"kind" binding:"required,oneof=provincial regional departamental barrial"`
	ParentZoneID *uint  `json:"parent_zone_id"`
	IsActive     *bool  `json:"is_active"`
}

type UpdateZoneRequest struct {
	Name         *string `json:"name"`
	Kind         *string `json:"kind" binding:"omitempty,oneof=provincial regional departamental barrial"`
	ParentZoneID *uint   `json:"parent_zone_id"`
	IsActive     *bool   `json:"is_active"`
}

type DeliveryRate struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	ZoneID                uint           `gorm:"not null;uniqueIndex:uq_rates_zone_branch" json:"zone_id"`
	OriginBranchID        uint           `gorm:"not null;uniqueIndex:uq_rates_zone_branch" json:"origin_branch_id"`
	Cost                  float64        `gorm:"type:decimal(12,2);not null" json:"cost"`
	EtaMinDays            int            `gorm:"not null" json:"eta_min_days"`
	EtaMaxDays            int            `gorm:"not null" json:"eta_max_days"`
	FreeShippingThreshold *float64       `gorm:"type:decimal(12,2)" json:"free_shipping_threshold,omitempty"`
	IsActive              bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DeliveryRate) TableName() string { return "delivery_rates" }

type CreateRateRequest struct {
	ZoneID                uint     `json:"zone_id" binding:"required"`
	OriginBranchID        uint     `json:"origin_branch_id" binding:"required"`
	Cost                  float64  `json:"cost" binding:"gte=0"`
	EtaMinDays            int      `json:"eta_min_days" binding:"gte=0"`
	EtaMaxDays            int      `json:"eta_max_days" binding:"gte=0"`
	FreeShippingThreshold *float64 `json:"free_shipping_threshold"`
	IsActive              *bool    `json:"is_active"`
}

type UpdateRateRequest struct {
	Cost                  *float64 `json:"cost" binding:"omitempty,gte=0"`
	EtaMinDays            *int     `json:"eta_min_days" binding:"omitempty,gte=0"`
	EtaMaxDays            *int     `json:"eta_max_days" binding:"omitempty,gte=0"`
	FreeShippingThreshold *float64 `json:"free_shipping_threshold"`
	IsActive              *bool    `json:"is_active"`
}

type PostalCodeZone struct {
	PostalCode string    `gorm:"primaryKey;type:varchar(16)" json:"postal_code"`
	ZoneID     uint      `gorm:"not null;index" json:"zone_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (PostalCodeZone) TableName() string { return "postal_code_zones" }

type UpsertPostalCodeRequest struct {
	PostalCode string `json:"postal_code" binding:"required,max=16"`
	ZoneID     uint   `json:"zone_id" binding:"required"`
}

type BulkPostalCodeEntry struct {
	PostalCode string `json:"postal_code" binding:"required,max=16"`
	ZoneID     uint   `json:"zone_id" binding:"required"`
}

type BulkPostalCodeRequest struct {
	Entries []BulkPostalCodeEntry `json:"entries" binding:"required,min=1,dive"`
}

// ===== Shipping quote =====

type QuoteItem struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type ShippingQuoteRequest struct {
	PostalCode string      `json:"postal_code" binding:"required,max=16"`
	Subtotal   float64     `json:"subtotal" binding:"gte=0"`
	Items      []QuoteItem `json:"items" binding:"required,min=1,dive"`
}

type ShippingQuoteZone struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ShippingQuoteResponse struct {
	Zone                 ShippingQuoteZone `json:"zone"`
	OriginBranchID       uint              `json:"origin_branch_id"`
	OriginBranchName     string            `json:"origin_branch_name"`
	Cost                 float64           `json:"cost"`
	EtaMinDays           int               `json:"eta_min_days"`
	EtaMaxDays           int               `json:"eta_max_days"`
	FreeShippingApplied  bool              `json:"free_shipping_applied"`
	AmountForFreeShip    *float64          `json:"amount_for_free,omitempty"`
	InStock              bool              `json:"in_stock"`
	OutOfStockProductIDs []uint            `json:"out_of_stock_items"`
}
