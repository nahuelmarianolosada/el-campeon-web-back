package models

import "time"

type OrderReportItem struct {
	ID            uint      `json:"id"`
	OrderNumber   string    `json:"order_number"`
	Email         string    `json:"email"`
	Productos     string    `json:"productos"`
	CantidadItems int       `json:"cantidad_items"`
	Total         float64   `json:"total"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type LowStockProduct struct {
	ID       uint   `json:"id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Stock    int    `json:"stock"`
	Category string `json:"category"`
}

type DailyRevenue struct {
	Fecha           string  `json:"fecha"`
	CantidadOrdenes int     `json:"cantidad_ordenes"`
	Ingresos        float64 `json:"ingresos"`
}
