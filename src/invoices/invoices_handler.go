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

// GET /invoices/:id or /buildings/:id/invoices/:invoiceId
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	// Try invoiceId first (for building-scoped routes), then id (for legacy routes)
	idStr := c.Param("invoiceId")
	if idStr == "" {
		idStr = c.Param("id")
	}
	
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

	// Get invoice items and splits for full details
	invoiceItems, _ := h.service.GetInvoiceItemRepo().GetByInvoiceID(id)
	splits, _ := h.service.GetSplitRepo().GetByTransactionID(invoice.TransactionID)
	transaction, _ := h.service.GetTransactionRepo().GetByID(invoice.TransactionID)

	response := gin.H{
		"invoice":     invoice,
		"items":       invoiceItems,
		"splits":      splits,
		"transaction": transaction,
	}

	c.JSON(http.StatusOK, response)
}

// PUT /invoices/:id or /buildings/:id/invoices/:invoiceId
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	var req UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Get invoice ID from route parameter (could be "id" or "invoiceId" depending on route)
	invoiceIDStr := c.Param("invoiceId")
	if invoiceIDStr == "" {
		invoiceIDStr = c.Param("id")
	}
	
	id, err := strconv.Atoi(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Invoice ID"})
		return
	}
	req.ID = id

	// Get building_id from route if building-scoped
	buildingIDStr := c.Param("id") // This is the building ID in building-scoped routes
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

	// If building ID not found in route, try query parameter
	if req.BuildingID == 0 {
		buildingIDStr = c.Query("building_id")
		if buildingIDStr != "" {
			buildingID, err := strconv.Atoi(buildingIDStr)
			if err == nil {
				req.BuildingID = buildingID
			}
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

	response, err := h.service.UpdateInvoice(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

