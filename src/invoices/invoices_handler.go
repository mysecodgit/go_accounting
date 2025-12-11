package invoices

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	service *InvoiceService
}

func NewInvoiceHandler(service *InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{service: service}
}

// POST /invoices/preview or /buildings/:id/invoices/preview
func (h *InvoiceHandler) PreviewInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Get building_id from route if building-scoped
	buildingIDStr := c.Param("id")
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

	// Get user ID from context or header (you may need to adjust this based on your auth setup)
	userIDStr := c.GetHeader("User-ID")
	if userIDStr == "" {
		userIDStr = c.Query("user_id")
	}
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	preview, err := h.service.PreviewInvoice(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// POST /invoices or /buildings/:id/invoices
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Get building_id from route if building-scoped
	buildingIDStr := c.Param("id")
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

	// Get user ID from context or header
	userIDStr := c.GetHeader("User-ID")
	if userIDStr == "" {
		userIDStr = c.Query("user_id")
	}
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	response, err := h.service.CreateInvoice(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /invoices
func (h *InvoiceHandler) GetInvoices(c *gin.Context) {
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		buildingIDStr = c.Query("building_id")
	}

	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Building ID"})
		return
	}

	invoices, err := h.service.GetInvoiceRepo().GetByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

// GET /invoices/:id
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	invoice, err := h.service.GetInvoiceRepo().GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

