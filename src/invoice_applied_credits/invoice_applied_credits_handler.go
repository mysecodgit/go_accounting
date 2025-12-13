package invoice_applied_credits

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceAppliedCreditHandler struct {
	service *InvoiceAppliedCreditService
}

func NewInvoiceAppliedCreditHandler(service *InvoiceAppliedCreditService) *InvoiceAppliedCreditHandler {
	return &InvoiceAppliedCreditHandler{service: service}
}

// GET /invoices/:invoiceId/available-credits or /buildings/:id/invoices/:invoiceId/available-credits
func (h *InvoiceAppliedCreditHandler) GetAvailableCredits(c *gin.Context) {
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

	response, err := h.service.GetAvailableCreditsForInvoice(invoiceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// POST /invoices/:invoiceId/apply-credit or /buildings/:id/invoices/:invoiceId/apply-credit
func (h *InvoiceAppliedCreditHandler) ApplyCreditToInvoice(c *gin.Context) {
	var req CreateInvoiceAppliedCreditRequest
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

	response, err := h.service.ApplyCreditToInvoice(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /invoices/:invoiceId/applied-credits or /buildings/:id/invoices/:invoiceId/applied-credits
func (h *InvoiceAppliedCreditHandler) GetAppliedCredits(c *gin.Context) {
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

	appliedCredits, err := h.service.GetAppliedCreditsByInvoiceID(invoiceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appliedCredits)
}

// POST /invoices/:invoiceId/preview-apply-credit or /buildings/:id/invoices/:invoiceId/preview-apply-credit
func (h *InvoiceAppliedCreditHandler) PreviewApplyCredit(c *gin.Context) {
	var req CreateInvoiceAppliedCreditRequest
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

	preview, err := h.service.PreviewApplyCredit(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// DELETE /invoice-applied-credits/:appliedCreditId or /buildings/:id/invoice-applied-credits/:appliedCreditId
func (h *InvoiceAppliedCreditHandler) DeleteAppliedCredit(c *gin.Context) {
	appliedCreditIDStr := c.Param("appliedCreditId")
	if appliedCreditIDStr == "" {
		// Try alternative param name
		appliedCreditIDStr = c.Param("id")
	}
	if appliedCreditIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Applied Credit ID is required"})
		return
	}

	appliedCreditID, err := strconv.Atoi(appliedCreditIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Applied Credit ID"})
		return
	}

	err = h.service.DeleteAppliedCredit(appliedCreditID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Applied credit deleted successfully"})
}
