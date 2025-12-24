package invoice_applied_discounts

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceAppliedDiscountHandler struct {
	service *InvoiceAppliedDiscountService
}

func NewInvoiceAppliedDiscountHandler(service *InvoiceAppliedDiscountService) *InvoiceAppliedDiscountHandler {
	return &InvoiceAppliedDiscountHandler{service: service}
}

// POST /invoices/:invoiceId/apply-discount or /buildings/:id/invoices/:invoiceId/apply-discount
func (h *InvoiceAppliedDiscountHandler) ApplyDiscountToInvoice(c *gin.Context) {
	var req CreateInvoiceAppliedDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Get invoice ID from route
	invoiceIDStr := c.Param("invoiceId")
	if invoiceIDStr == "" {
		invoiceIDStr = c.Param("id")
	}
	if invoiceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice ID is required"})
		return
	}

	invoiceID, err := strconv.Atoi(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}
	req.InvoiceID = invoiceID

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

	response, err := h.service.ApplyDiscountToInvoice(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /invoices/:invoiceId/applied-discounts or /buildings/:id/invoices/:invoiceId/applied-discounts
func (h *InvoiceAppliedDiscountHandler) GetAppliedDiscounts(c *gin.Context) {
	invoiceIDStr := c.Param("invoiceId")
	if invoiceIDStr == "" {
		invoiceIDStr = c.Param("id")
	}
	if invoiceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice ID is required"})
		return
	}

	invoiceID, err := strconv.Atoi(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	appliedDiscounts, err := h.service.GetAppliedDiscountsByInvoiceID(invoiceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appliedDiscounts)
}

// POST /invoices/:invoiceId/preview-apply-discount or /buildings/:id/invoices/:invoiceId/preview-apply-discount
func (h *InvoiceAppliedDiscountHandler) PreviewApplyDiscount(c *gin.Context) {
	var req CreateInvoiceAppliedDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Get invoice ID from route
	invoiceIDStr := c.Param("invoiceId")
	if invoiceIDStr == "" {
		invoiceIDStr = c.Param("id")
	}
	if invoiceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice ID is required"})
		return
	}

	invoiceID, err := strconv.Atoi(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}
	req.InvoiceID = invoiceID

	preview, err := h.service.PreviewApplyDiscount(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// DELETE /invoice-applied-discounts/:appliedDiscountId or /buildings/:id/invoice-applied-discounts/:appliedDiscountId
func (h *InvoiceAppliedDiscountHandler) DeleteAppliedDiscount(c *gin.Context) {
	appliedDiscountIDStr := c.Param("appliedDiscountId")
	if appliedDiscountIDStr == "" {
		// Try alternative param name
		appliedDiscountIDStr = c.Param("id")
	}
	if appliedDiscountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Applied Discount ID is required"})
		return
	}

	appliedDiscountID, err := strconv.Atoi(appliedDiscountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Applied Discount ID"})
		return
	}

	err = h.service.DeleteAppliedDiscount(appliedDiscountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Applied discount deleted successfully"})
}

