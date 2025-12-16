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

	// Get filter parameters from query string
	var startDate, endDate, status *string
	var peopleID *int

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate = &startDateStr
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate = &endDateStr
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}
	if peopleIDStr := c.Query("people_id"); peopleIDStr != "" {
		if pid, err := strconv.Atoi(peopleIDStr); err == nil {
			peopleID = &pid
		}
	}

	payments, err := h.service.GetPaymentRepo().GetByBuildingIDWithFilters(buildingID, startDate, endDate, peopleID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// GET /invoice-payments/:id or /buildings/:id/invoice-payments/:paymentId
func (h *InvoicePaymentHandler) GetInvoicePayment(c *gin.Context) {
	// For building-scoped routes, use paymentId; for legacy routes, use id
	idStr := c.Param("paymentId")
	if idStr == "" {
		idStr = c.Param("id")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Payment ID"})
		return
	}

	response, err := h.service.GetInvoicePaymentWithDetails(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
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

// POST /buildings/:id/invoice-payments/preview
func (h *InvoicePaymentHandler) PreviewInvoicePayment(c *gin.Context) {
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

	preview, err := h.service.PreviewInvoicePayment(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// PUT /buildings/:id/invoice-payments/:paymentId
func (h *InvoicePaymentHandler) UpdateInvoicePayment(c *gin.Context) {
	paymentIDStr := c.Param("paymentId")
	if paymentIDStr == "" {
		paymentIDStr = c.Param("id")
	}

	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Payment ID"})
		return
	}

	var req UpdateInvoicePaymentRequest
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

	response, err := h.service.UpdateInvoicePayment(paymentID, req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
