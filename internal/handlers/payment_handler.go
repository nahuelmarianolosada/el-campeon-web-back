package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/payment"
)

type PaymentHandler struct {
	paymentService payment.PaymentService
}

func NewPaymentHandler(paymentService payment.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePayment crea un nuevo pago
// @Summary Crear un pago
// @Tags Pagos
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.CreatePaymentRequest true "Datos del pago"
// @Success 201 {object} models.PaymentResponse
// @Failure 400 {object} gin.H
// @Router /api/payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	ctx := c.Request.Context()
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar que el usuario es propietario de la orden
	// En una implementación completa, consultaríamos la orden

	payment, err := h.paymentService.CreatePayment(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// GetPayment obtiene un pago por ID
// @Summary Obtener pago por ID
// @Tags Pagos
// @Security Bearer
// @Produce json
// @Param id path int true "ID del pago"
// @Success 200 {object} models.PaymentResponse
// @Failure 404 {object} gin.H
// @Router /api/payments/:id [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	payment, err := h.paymentService.GetPaymentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetMyPayments obtiene los pagos del usuario autenticado
// @Summary Obtener mis pagos
// @Tags Pagos
// @Security Bearer
// @Produce json
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.PaymentResponse
// @Router /api/payments/my [get]
func (h *PaymentHandler) GetMyPayments(c *gin.Context) {
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

	payments, err := h.paymentService.GetPaymentsByUserID(userID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   payments,
		"limit":  limit,
		"offset": offset,
	})
}

// GetPaymentByOrderID obtiene el pago de una orden
// @Summary Obtener pago de una orden
// @Tags Pagos
// @Security Bearer
// @Produce json
// @Param orderId path int true "ID de la orden"
// @Success 200 {object} models.PaymentResponse
// @Failure 404 {object} gin.H
// @Router /api/payments/order/:orderId [get]
func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("orderId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	payment, err := h.paymentService.GetPaymentByOrderID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// UpdatePaymentStatus actualiza el estado de un pago (solo ADMIN)
// @Summary Actualizar estado de pago
// @Tags Pagos
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID del pago"
// @Param request body gin.H true "Estado del pago"
// @Success 200 {object} models.PaymentResponse
// @Failure 400 {object} gin.H
// @Router /api/payments/:id/status [put]
func (h *PaymentHandler) UpdatePaymentStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentService.UpdatePaymentStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// ListAllPayments lista todos los pagos (solo ADMIN)
// @Summary Listar todos los pagos
// @Tags Pagos
// @Security Bearer
// @Produce json
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.PaymentResponse
// @Router /api/payments [get]
func (h *PaymentHandler) ListAllPayments(c *gin.Context) {
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

	payments, err := h.paymentService.ListAllPayments(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   payments,
		"limit":  limit,
		"offset": offset,
	})
}

// MercadopagoWebhook maneja webhooks de MercadoPago
// @Summary Webhook de MercadoPago
// @Tags Pagos
// @Accept json
// @Produce json
// @Param request body models.MercadopagoWebhookRequest true "Webhook de MercadoPago"
// @Header 200 {string} X-Signature "Firma del webhook"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /webhooks/mercadopago [post]
func (h *PaymentHandler) MercadopagoWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	var webhook models.MercadopagoWebhookRequest

	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obtener el header de firma
	xSignature := c.GetHeader("X-Signature")
	if xSignature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-Signature header"})
		return
	}

	// Procesar el webhook
	if err := h.paymentService.ProcessMercadopagoWebhook(ctx, &webhook, xSignature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook processed"})
}
