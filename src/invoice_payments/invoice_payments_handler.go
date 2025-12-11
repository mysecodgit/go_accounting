package invoice_payments

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoicePaymentHandler struct {
	service *InvoicePaymentService
}

func NewInvoicePaymentHandler(service *InvoicePaymentService) *InvoicePaymentHandler {
	return &InvoicePaymentHandler{service: service}
}

// POST /invoice-payments or /buildings/:id/invoice-payments
func (h *InvoicePaymentHandler) CreateInvoicePayment(c *gin.Context) {
	var req CreateInvoicePaymentRequest
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

	response, err := h.service.CreateInvoicePayment(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /invoice-payments
func (h *InvoicePaymentHandler) GetInvoicePayments(c *gin.Context) {
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		buildingIDStr = c.Query("building_id")
	}

	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Building ID"})
		return
	}

	payments, err := h.service.GetPaymentRepo().GetByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// GET /invoice-payments/:id
func (h *InvoicePaymentHandler) GetInvoicePayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	payment, err := h.service.GetPaymentRepo().GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GET /invoices/:id/payments or /buildings/:id/invoices/:invoiceId/payments
func (h *InvoicePaymentHandler) GetPaymentsByInvoice(c *gin.Context) {
	// Try invoiceId first (for building-scoped routes), then id (for legacy routes)
	invoiceIDStr := c.Param("invoiceId")
	if invoiceIDStr == "" {
		invoiceIDStr = c.Param("id")
	}

	invoiceID, err := strconv.Atoi(invoiceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Invoice ID"})
		return
	}

	payments, err := h.service.GetPaymentRepo().GetByInvoiceID(invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}
