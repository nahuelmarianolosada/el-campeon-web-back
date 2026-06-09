package models

import (
	"time"

	"gorm.io/gorm"
)

type Branch struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Code          string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"code"`
	Name          string         `gorm:"type:varchar(120);not null" json:"name"`
	Address       string         `gorm:"type:varchar(255);not null" json:"address"`
	Lat           *float64       `gorm:"type:decimal(10,7)" json:"lat,omitempty"`
	Lng           *float64       `gorm:"type:decimal(10,7)" json:"lng,omitempty"`
	IsPickupPoint bool           `gorm:"not null;default:true" json:"is_pickup_point"`
	IsActive      bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Branch) TableName() string { return "branches" }

type CreateBranchRequest struct {
	Code          string   `json:"code" binding:"required,max=32"`
	Name          string   `json:"name" binding:"required,max=120"`
	Address       string   `json:"address" binding:"required,max=255"`
	Lat           *float64 `json:"lat"`
	Lng           *float64 `json:"lng"`
	IsPickupPoint *bool    `json:"is_pickup_point"`
	IsActive      *bool    `json:"is_active"`
}

type UpdateBranchRequest struct {
	Name          *string  `json:"name"`
	Address       *string  `json:"address"`
	Lat           *float64 `json:"lat"`
	Lng           *float64 `json:"lng"`
	IsPickupPoint *bool    `json:"is_pickup_point"`
	IsActive      *bool    `json:"is_active"`
}

type BranchResponse struct {
	ID            uint     `json:"id"`
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Lat           *float64 `json:"lat,omitempty"`
	Lng           *float64 `json:"lng,omitempty"`
	IsPickupPoint bool     `json:"is_pickup_point"`
	IsActive      bool     `json:"is_active"`
}

func (b *Branch) ToResponse() BranchResponse {
	return BranchResponse{
		ID:            b.ID,
		Code:          b.Code,
		Name:          b.Name,
		Address:       b.Address,
		Lat:           b.Lat,
		Lng:           b.Lng,
		IsPickupPoint: b.IsPickupPoint,
		IsActive:      b.IsActive,
	}
}

type ProductBranchStock struct {
	ProductID uint      `gorm:"primaryKey" json:"product_id"`
	BranchID  uint      `gorm:"primaryKey" json:"branch_id"`
	Stock     int       `gorm:"not null;default:0" json:"stock"`
	Reserved  int       `gorm:"not null;default:0" json:"reserved"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ProductBranchStock) TableName() string { return "product_branch_stock" }

type ProductBranchStockResponse struct {
	ProductID  uint   `json:"product_id"`
	BranchID   uint   `json:"branch_id"`
	BranchCode string `json:"branch_code,omitempty"`
	BranchName string `json:"branch_name,omitempty"`
	Stock      int    `json:"stock"`
	Reserved   int    `json:"reserved"`
	Available  int    `json:"available"`
}

type UpdateBranchStockRequest struct {
	BranchID uint `json:"branch_id" binding:"required"`
	Stock    int  `json:"stock" binding:"gte=0"`
}
