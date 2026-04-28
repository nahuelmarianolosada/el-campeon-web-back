package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint       `gorm:"primaryKey"`
	Email       string     `gorm:"type:varchar(255);uniqueIndex:idx_users_email;not null"`
	FirstName   string     `gorm:"type:varchar(255);not null"`
	LastName    string     `gorm:"type:varchar(255);not null"`
	Password    string     `gorm:"type:varchar(255);not null"`
	Phone       string     `gorm:"type:varchar(20)"`
	Address     string     `gorm:"type:text"`
	City        string     `gorm:"type:varchar(100)"`
	PostalCode  string     `gorm:"type:varchar(20)"`
	Country     string     `gorm:"type:varchar(100)"`
	Role        string     `gorm:"type:enum('USER','ADMIN');default:'USER'"`
	IsActive    bool       `gorm:"default:true"`
	IsBulkBuyer bool       `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type UserResponse struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	PostalCode  string    `json:"postal_code"`
	Country     string    `json:"country"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	IsBulkBuyer bool      `json:"is_bulk_buyer"`
	CreatedAt   time.Time `json:"created_at"`
}

// RegisterRequest estructura para registro de usuario
type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Password   string `json:"password" binding:"required,min=8"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// LoginRequest estructura para login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse estructura para respuesta de autenticación
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
	ExpiresIn    int64        `json:"expires_in"`
}
