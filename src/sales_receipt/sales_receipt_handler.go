package sales_receipt

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SalesReceiptHandler struct {
	service *SalesReceiptService
}

func NewSalesReceiptHandler(service *SalesReceiptService) *SalesReceiptHandler {
	return &SalesReceiptHandler{service: service}
}

// POST /sales-receipts/preview or /buildings/:id/sales-receipts/preview
func (h *SalesReceiptHandler) PreviewSalesReceipt(c *gin.Context) {
	var req CreateSalesReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	buildingIDStr := c.Param("id")
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

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

	preview, err := h.service.PreviewSalesReceipt(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// POST /sales-receipts or /buildings/:id/sales-receipts
func (h *SalesReceiptHandler) CreateSalesReceipt(c *gin.Context) {
	var req CreateSalesReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	buildingIDStr := c.Param("id")
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

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

	response, err := h.service.CreateSalesReceipt(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /sales-receipts
func (h *SalesReceiptHandler) GetSalesReceipts(c *gin.Context) {
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		buildingIDStr = c.Query("building_id")
	}

	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Building ID"})
		return
	}

	receipts, err := h.service.GetReceiptRepo().GetByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, receipts)
}

// GET /sales-receipts/:id or /buildings/:id/sales-receipts/:receiptId
func (h *SalesReceiptHandler) GetSalesReceipt(c *gin.Context) {
	// Try receiptId first (for building-scoped routes), then id (for legacy routes)
	idStr := c.Param("receiptId")
	if idStr == "" {
		idStr = c.Param("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	receipt, err := h.service.GetReceiptRepo().GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get receipt items and splits for full details
	receiptItems, _ := h.service.GetReceiptItemRepo().GetByReceiptID(id)
	splits, _ := h.service.GetSplitRepo().GetByTransactionID(receipt.TransactionID)
	transaction, _ := h.service.GetTransactionRepo().GetByID(receipt.TransactionID)

	c.JSON(http.StatusOK, SalesReceiptResponse{
		Receipt:     receipt,
		Items:       receiptItems,
		Splits:      splits,
		Transaction: transaction,
	})
}

// PUT /sales-receipts/:id or /buildings/:id/sales-receipts/:receiptId
func (h *SalesReceiptHandler) UpdateSalesReceipt(c *gin.Context) {
	// Try receiptId first (for building-scoped routes), then id (for legacy routes)
	idStr := c.Param("receiptId")
	if idStr == "" {
		idStr = c.Param("id")
	}

	receiptID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Receipt ID"})
		return
	}

	var req UpdateSalesReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	req.ID = receiptID // Set ID from URL parameter

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

	response, err := h.service.UpdateSalesReceipt(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
