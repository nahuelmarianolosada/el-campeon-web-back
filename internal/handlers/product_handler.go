package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/product"
)

type ProductHandler struct {
	productService product.ProductService
}

func NewProductHandler(productService product.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct crea un nuevo producto (solo ADMIN)
// @Summary Crear nuevo producto
// @Tags Productos
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body models.CreateProductRequest true "Datos del producto"
// @Success 201 {object} models.ProductResponse
// @Failure 400 {object} gin.H
// @Router /api/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.CreateProduct(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProduct obtiene un producto por ID
// @Summary Obtener producto por ID
// @Tags Productos
// @Produce json
// @Param id path int true "ID del producto"
// @Success 200 {object} models.ProductResponse
// @Failure 404 {object} gin.H
// @Router /api/products/:id [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.productService.GetProductByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetProductBySKU obtiene un producto por SKU
// @Summary Obtener producto por SKU
// @Tags Productos
// @Produce json
// @Param sku query string true "SKU del producto"
// @Success 200 {object} models.ProductResponse
// @Failure 404 {object} gin.H
// @Router /api/products/sku [get]
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Query("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku parameter required"})
		return
	}

	product, err := h.productService.GetProductBySKU(sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct actualiza un producto (solo ADMIN)
// @Summary Actualizar producto
// @Tags Productos
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "ID del producto"
// @Param request body models.UpdateProductRequest true "Datos a actualizar"
// @Success 200 {object} models.ProductResponse
// @Failure 400 {object} gin.H
// @Router /api/products/:id [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.UpdateProduct(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct elimina un producto (solo ADMIN)
// @Summary Eliminar producto
// @Tags Productos
// @Security Bearer
// @Param id path int true "ID del producto"
// @Success 204
// @Failure 404 {object} gin.H
// @Router /api/products/:id [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.productService.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListProducts lista todos los productos
// @Summary Listar productos
// @Tags Productos
// @Produce json
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.ProductResponse
// @Router /api/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
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

	products, err := h.productService.ListActiveProducts(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   products,
		"limit":  limit,
		"offset": offset,
	})
}

// ImportProducts importa productos desde un archivo .xlsx o .csv (solo ADMIN).
// Recibe el archivo en el campo multipart "file" y un flag "dry_run" (string "true"/"false",
// default "true") que controla si se aplica o sólo se previsualiza.
// @Summary Importar productos desde Excel/CSV
// @Tags Productos
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Archivo .xlsx o .csv"
// @Param dry_run formData string false "true para previsualizar sin aplicar (default true)"
// @Success 200 {object} product.ImportResult
// @Failure 400 {object} gin.H
// @Router /api/products/import [post]
func (h *ProductHandler) ImportProducts(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "archivo requerido en el campo 'file'"})
		return
	}

	dryRun := true
	if v := strings.TrimSpace(strings.ToLower(c.PostForm("dry_run"))); v != "" {
		dryRun = v != "false" && v != "0"
	}

	const maxSize = 10 << 20 // 10 MiB
	if fileHeader.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "archivo demasiado grande (máximo 10 MB)"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo abrir el archivo"})
		return
	}
	defer file.Close()

	result, err := h.productService.ImportProducts(file, fileHeader.Filename, dryRun)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ListProductsByCategory lista productos por categoría
// @Summary Listar productos por categoría
// @Tags Productos
// @Produce json
// @Param category query string true "Categoría"
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.ProductResponse
// @Router /api/products/category/:category [get]
func (h *ProductHandler) ListProductsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category parameter required"})
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

	products, err := h.productService.ListProductsByCategory(category, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     products,
		"category": category,
		"limit":    limit,
		"offset":   offset,
	})
}
