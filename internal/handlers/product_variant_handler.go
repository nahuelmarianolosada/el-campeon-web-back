package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/product/variant"
)

type ProductVariantHandler struct {
	variantService variant.ProductVariantService
}

func NewProductVariantHandler(variantService variant.ProductVariantService) *ProductVariantHandler {
	return &ProductVariantHandler{
		variantService: variantService,
	}
}

// Variant endpoints

// CreateProductVariant crea una nueva variante de producto (solo ADMIN)
// @Summary Crear variante de producto
// @Tags Variantes
// @Security Bearer
// @Accept json
// @Produce json
// @Param productId path int true "ID del producto"
// @Param request body models.CreateProductVariantRequest true "Datos de la variante"
// @Success 201 {object} models.ProductVariantResponse
// @Failure 400 {object} gin.H
// @Router /api/products/:productId/variants [post]
func (h *ProductVariantHandler) CreateProductVariant(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req models.CreateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.variantService.CreateVariant(uint(productID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetProductVariants obtiene todas las variantes de un producto
// @Summary Obtener variantes del producto
// @Tags Variantes
// @Produce json
// @Param productId path int true "ID del producto"
// @Success 200 {array} models.ProductVariantResponse
// @Failure 404 {object} gin.H
// @Router /api/products/:productId/variants [get]
func (h *ProductVariantHandler) GetProductVariants(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	variants, err := h.variantService.GetProductVariants(uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variants)
}

// GetVariant obtiene una variante específica
// @Summary Obtener variante
// @Tags Variantes
// @Produce json
// @Param variantId path int true "ID de la variante"
// @Success 200 {object} models.ProductVariantResponse
// @Failure 404 {object} gin.H
// @Router /api/variants/:variantId [get]
func (h *ProductVariantHandler) GetVariant(c *gin.Context) {
	variantID, err := strconv.ParseUint(c.Param("variantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant id"})
		return
	}

	variantResp, err := h.variantService.GetVariant(uint(variantID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variantResp)
}

// UpdateProductVariant actualiza una variante (solo ADMIN)
// @Summary Actualizar variante
// @Tags Variantes
// @Security Bearer
// @Accept json
// @Produce json
// @Param variantId path int true "ID de la variante"
// @Param request body models.UpdateProductVariantRequest true "Datos de la variante"
// @Success 200 {object} models.ProductVariantResponse
// @Failure 400 {object} gin.H
// @Router /api/variants/:variantId [put]
func (h *ProductVariantHandler) UpdateProductVariant(c *gin.Context) {
	variantID, err := strconv.ParseUint(c.Param("variantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant id"})
		return
	}

	var req models.UpdateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variantResp, err := h.variantService.UpdateVariant(uint(variantID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variantResp)
}

// DeleteProductVariant elimina una variante (solo ADMIN)
// @Summary Eliminar variante
// @Tags Variantes
// @Security Bearer
// @Param variantId path int true "ID de la variante"
// @Success 204
// @Failure 404 {object} gin.H
// @Router /api/variants/:variantId [delete]
func (h *ProductVariantHandler) DeleteProductVariant(c *gin.Context) {
	variantID, err := strconv.ParseUint(c.Param("variantId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant id"})
		return
	}

	if err := h.variantService.DeleteVariant(uint(variantID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Variant Combination endpoints

// CreateVariantCombination crea una nueva combinación de variantes (solo ADMIN)
// @Summary Crear combinación de variantes
// @Tags Combinaciones de Variantes
// @Security Bearer
// @Accept json
// @Produce json
// @Param productId path int true "ID del producto"
// @Param request body models.CreateProductVariantCombinationRequest true "Datos de la combinación"
// @Success 201 {object} models.ProductVariantCombinationResponse
// @Failure 400 {object} gin.H
// @Router /api/products/:productId/variant-combinations [post]
func (h *ProductVariantHandler) CreateVariantCombination(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req models.CreateProductVariantCombinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.variantService.CreateVariantCombination(uint(productID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetVariantCombination obtiene una combinación específica
// @Summary Obtener combinación de variantes
// @Tags Combinaciones de Variantes
// @Produce json
// @Param combinationId path int true "ID de la combinación"
// @Success 200 {object} models.ProductVariantCombinationResponse
// @Failure 404 {object} gin.H
// @Router /api/variant-combinations/:combinationId [get]
func (h *ProductVariantHandler) GetVariantCombination(c *gin.Context) {
	combinationID, err := strconv.ParseUint(c.Param("combinationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid combination id"})
		return
	}

	combination, err := h.variantService.GetVariantCombination(uint(combinationID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, combination)
}

// GetVariantCombinationBySKU obtiene una combinación por SKU
// @Summary Obtener combinación por SKU
// @Tags Combinaciones de Variantes
// @Produce json
// @Param sku query string true "SKU de la combinación"
// @Success 200 {object} models.ProductVariantCombinationResponse
// @Failure 404 {object} gin.H
// @Router /api/variant-combinations/sku [get]
func (h *ProductVariantHandler) GetVariantCombinationBySKU(c *gin.Context) {
	sku := c.Query("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku parameter required"})
		return
	}

	combination, err := h.variantService.GetVariantCombinationBySKU(sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, combination)
}

// GetProductVariantCombinations obtiene todas las combinaciones de un producto
// @Summary Obtener combinaciones de variantes del producto
// @Tags Combinaciones de Variantes
// @Produce json
// @Param productId path int true "ID del producto"
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.ProductVariantCombinationResponse
// @Failure 404 {object} gin.H
// @Router /api/products/:productId/variant-combinations [get]
func (h *ProductVariantHandler) GetProductVariantCombinations(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
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

	combinations, err := h.variantService.GetProductVariantCombinations(uint(productID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   combinations,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateVariantCombination actualiza una combinación (solo ADMIN)
// @Summary Actualizar combinación de variantes
// @Tags Combinaciones de Variantes
// @Security Bearer
// @Accept json
// @Produce json
// @Param combinationId path int true "ID de la combinación"
// @Param request body models.UpdateProductVariantCombinationRequest true "Datos de la combinación"
// @Success 200 {object} models.ProductVariantCombinationResponse
// @Failure 400 {object} gin.H
// @Router /api/variant-combinations/:combinationId [put]
func (h *ProductVariantHandler) UpdateVariantCombination(c *gin.Context) {
	combinationID, err := strconv.ParseUint(c.Param("combinationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid combination id"})
		return
	}

	var req models.UpdateProductVariantCombinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	combination, err := h.variantService.UpdateVariantCombination(uint(combinationID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, combination)
}

// DeleteVariantCombination elimina una combinación (solo ADMIN)
// @Summary Eliminar combinación de variantes
// @Tags Combinaciones de Variantes
// @Security Bearer
// @Param combinationId path int true "ID de la combinación"
// @Success 204
// @Failure 404 {object} gin.H
// @Router /api/variant-combinations/:combinationId [delete]
func (h *ProductVariantHandler) DeleteVariantCombination(c *gin.Context) {
	combinationID, err := strconv.ParseUint(c.Param("combinationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid combination id"})
		return
	}

	if err := h.variantService.DeleteVariantCombination(uint(combinationID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
