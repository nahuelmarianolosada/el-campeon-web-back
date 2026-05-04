package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/cart"
)

type CartHandler struct {
	cartService cart.CartService
}

func NewCartHandler(cartService cart.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// AddToCart agrega un producto al carrito
// @Summary Agregar producto al carrito
// @Tags Carrito
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.AddToCartRequest true "Producto a agregar"
// @Success 201 {object} gin.H
// @Failure 400 {object} gin.H
// @Router /api/cart/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isBulkBuyer := false
	if role, exists := c.Get("role"); exists {
		// En una política real, buscaríamos el campo IsBulkBuyer del usuario
		// Por ahora simulamos
		_ = role
	}

	err := h.cartService.AddToCart(userID.(uint), &req, isBulkBuyer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "item added to cart"})
}

// GetCart obtiene el carrito del usuario
// @Summary Obtener carrito del usuario
// @Tags Carrito
// @Security Bearer
// @Produce json
// @Success 200 {object} models.CartResponse
// @Failure 404 {object} gin.H
// @Router /api/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	cart, err := h.cartService.GetCart(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// UpdateCartItem actualiza la cantidad de un item en el carrito
// @Summary Actualizar cantidad de item en carrito
// @Tags Carrito
// @Security Bearer
// @Accept json
// @Produce json
// @Param itemId path int true "ID del item del carrito"
// @Param request body models.UpdateCartItemRequest true "Nueva cantidad"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Router /api/cart/items/:itemId [put]
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.cartService.UpdateCartItem(userID.(uint), uint(itemID), req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cart item updated"})
}

// RemoveFromCart elimina un item del carrito
// @Summary Eliminar item del carrito
// @Tags Carrito
// @Security Bearer
// @Param itemId path int true "ID del item del carrito"
// @Success 204
// @Failure 400 {object} gin.H
// @Router /api/cart/items/:itemId [delete]
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	if err := h.cartService.RemoveFromCart(userID.(uint), uint(itemID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ClearCart vacía el carrito del usuario
// @Summary Vaciar carrito
// @Tags Carrito
// @Security Bearer
// @Success 204
// @Failure 400 {object} gin.H
// @Router /api/cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	if err := h.cartService.ClearCart(userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetCartTotal calcula el total del carrito
// @Summary Calcular total del carrito
// @Tags Carrito
// @Security Bearer
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/cart/total [get]
func (h *CartHandler) GetCartTotal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	total, err := h.cartService.CalculateCartTotal(userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total})
}
