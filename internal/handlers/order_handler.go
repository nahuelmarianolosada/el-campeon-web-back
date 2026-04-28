package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder crea una nueva orden a partir del carrito del usuario
// @Summary Crear una orden
// @Tags Órdenes
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.CreateOrderRequest true "Datos de la orden"
// @Success 201 {object} models.OrderResponse
// @Failure 400 {object} gin.H
// @Router /api/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrder obtiene una orden por ID
// @Summary Obtener orden por ID
// @Tags Órdenes
// @Security Bearer
// @Produce json
// @Param id path int true "ID de la orden"
// @Success 200 {object} models.OrderResponse
// @Failure 404 {object} gin.H
// @Router /api/orders/:id [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.orderService.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetMyOrders obtiene las órdenes del usuario autenticado
// @Summary Obtener mis órdenes
// @Tags Órdenes
// @Security Bearer
// @Produce json
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.OrderResponse
// @Router /api/orders/my [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	orders, err := h.orderService.GetOrdersByUserID(userID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   orders,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateOrderStatus actualiza el estado de una orden (solo ADMIN)
// @Summary Actualizar estado de orden
// @Tags Órdenes
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID de la orden"
// @Param request body gin.H true "Estado a actualizar"
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} gin.H
// @Router /api/orders/:id/status [put]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListAllOrders lista todas las órdenes (solo ADMIN)
// @Summary Listar todas las órdenes
// @Tags Órdenes
// @Security Bearer
// @Produce json
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.OrderResponse
// @Router /api/orders [get]
func (h *OrderHandler) ListAllOrders(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	orders, err := h.orderService.ListAllOrders(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   orders,
		"limit":  limit,
		"offset": offset,
	})
}

