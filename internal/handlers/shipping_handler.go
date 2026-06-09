package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/models"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/services/shipping"
)

type ShippingHandler struct {
	svc shipping.ShippingService
}

func NewShippingHandler(svc shipping.ShippingService) *ShippingHandler {
	return &ShippingHandler{svc: svc}
}

// ===== Endpoint público: cotización =====

// Quote cotiza el envío en base a CP, subtotal e items.
// @Summary Cotizar envío
// @Tags Envíos
// @Accept json
// @Produce json
// @Param request body models.ShippingQuoteRequest true "Datos de cotización"
// @Success 200 {object} models.ShippingQuoteResponse
// @Failure 400 {object} gin.H
// @Failure 422 {object} gin.H
// @Router /api/shipping/quote [post]
func (h *ShippingHandler) Quote(c *gin.Context) {
	var req models.ShippingQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Quote(&req)
	if err != nil {
		switch {
		case errors.Is(err, shipping.ErrPostalCodeNotCovered):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "POSTAL_CODE_NOT_COVERED"})
		case errors.Is(err, shipping.ErrNoBranchHasStock):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "NO_BRANCH_HAS_STOCK"})
		case errors.Is(err, shipping.ErrNoRateForZone):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "NO_RATE_FOR_ZONE"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ===== Branches =====

func (h *ShippingHandler) ListBranches(c *gin.Context) {
	onlyActive := c.DefaultQuery("active", "false") == "true"
	onlyPickup := c.DefaultQuery("is_pickup_point", "") == "true"
	branches, err := h.svc.ListBranches(onlyActive, onlyPickup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": branches})
}

func (h *ShippingHandler) GetBranch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}
	b, err := h.svc.GetBranch(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *ShippingHandler) CreateBranch(c *gin.Context) {
	var req models.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.svc.CreateBranch(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *ShippingHandler) UpdateBranch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}
	var req models.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.svc.UpdateBranch(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *ShippingHandler) DeleteBranch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}
	if err := h.svc.DeleteBranch(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== Zonas =====

func (h *ShippingHandler) ListZones(c *gin.Context) {
	onlyActive := c.DefaultQuery("active", "false") == "true"
	zones, err := h.svc.ListZones(onlyActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": zones})
}

func (h *ShippingHandler) CreateZone(c *gin.Context) {
	var req models.CreateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	z, err := h.svc.CreateZone(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, z)
}

func (h *ShippingHandler) UpdateZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	var req models.UpdateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	z, err := h.svc.UpdateZone(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, z)
}

func (h *ShippingHandler) DeleteZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	if err := h.svc.DeleteZone(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== Tarifas =====

func (h *ShippingHandler) ListRates(c *gin.Context) {
	var zoneID, branchID *uint
	if v := c.Query("zone_id"); v != "" {
		if parsed, err := strconv.ParseUint(v, 10, 32); err == nil {
			id := uint(parsed)
			zoneID = &id
		}
	}
	if v := c.Query("branch_id"); v != "" {
		if parsed, err := strconv.ParseUint(v, 10, 32); err == nil {
			id := uint(parsed)
			branchID = &id
		}
	}
	rates, err := h.svc.ListRates(zoneID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}

func (h *ShippingHandler) CreateRate(c *gin.Context) {
	var req models.CreateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.svc.CreateRate(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *ShippingHandler) UpdateRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rate id"})
		return
	}
	var req models.UpdateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.svc.UpdateRate(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *ShippingHandler) DeleteRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rate id"})
		return
	}
	if err := h.svc.DeleteRate(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== Códigos postales =====

func (h *ShippingHandler) ListPostalCodes(c *gin.Context) {
	var zoneID *uint
	if v := c.Query("zone_id"); v != "" {
		if parsed, err := strconv.ParseUint(v, 10, 32); err == nil {
			id := uint(parsed)
			zoneID = &id
		}
	}
	pcs, err := h.svc.ListPostalCodes(zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pcs})
}

func (h *ShippingHandler) UpsertPostalCode(c *gin.Context) {
	var req models.UpsertPostalCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpsertPostalCode(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ShippingHandler) BulkUpsertPostalCodes(c *gin.Context) {
	var req models.BulkPostalCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.BulkUpsertPostalCodes(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "count": len(req.Entries)})
}

func (h *ShippingHandler) DeletePostalCode(c *gin.Context) {
	pc := c.Param("postal_code")
	if pc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postal_code required"})
		return
	}
	if err := h.svc.DeletePostalCode(pc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== Stock por sucursal =====

func (h *ShippingHandler) GetProductStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	rows, err := h.svc.GetProductStock(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h *ShippingHandler) SetProductStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var req models.UpdateBranchStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetProductStock(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
