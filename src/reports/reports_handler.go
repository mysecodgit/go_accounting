package reports

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReportsHandler struct {
	service *ReportsService
}

func NewReportsHandler(service *ReportsService) *ReportsHandler {
	return &ReportsHandler{service: service}
}

// GET /reports/balance-sheet
func (h *ReportsHandler) GetBalanceSheet(c *gin.Context) {
	var req BalanceSheetRequest

	// Get building ID from route parameter (for building-scoped routes)
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		// Try query parameter (for legacy routes)
		buildingIDStr = c.Query("building_id")
	}
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

	// Get as_of_date from query
	req.AsOfDate = c.Query("as_of_date")

	if req.BuildingID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Building ID is required"})
		return
	}

	report, err := h.service.GetBalanceSheet(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/customers
func (h *ReportsHandler) GetCustomerReport(c *gin.Context) {
	var req CustomerVendorReportRequest

	// Get building ID from route parameter (for building-scoped routes)
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		// Try query parameter (for legacy routes)
		buildingIDStr = c.Query("building_id")
	}
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

	if peopleIDStr := c.Query("people_id"); peopleIDStr != "" {
		peopleID, err := strconv.Atoi(peopleIDStr)
		if err == nil {
			req.PeopleID = &peopleID
		}
	}

	if typeIDStr := c.Query("type_id"); typeIDStr != "" {
		typeID, err := strconv.Atoi(typeIDStr)
		if err == nil {
			req.TypeID = &typeID
		}
	}

	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")

	if req.BuildingID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Building ID is required"})
		return
	}

	if req.StartDate == "" || req.EndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start date and end date are required"})
		return
	}

	report, err := h.service.GetCustomerVendorReport(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/vendors
func (h *ReportsHandler) GetVendorReport(c *gin.Context) {
	var req CustomerVendorReportRequest

	// Get building ID from route parameter (for building-scoped routes)
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		// Try query parameter (for legacy routes)
		buildingIDStr = c.Query("building_id")
	}
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

	if peopleIDStr := c.Query("people_id"); peopleIDStr != "" {
		peopleID, err := strconv.Atoi(peopleIDStr)
		if err == nil {
			req.PeopleID = &peopleID
		}
	}

	if typeIDStr := c.Query("type_id"); typeIDStr != "" {
		typeID, err := strconv.Atoi(typeIDStr)
		if err == nil {
			req.TypeID = &typeID
		}
	}

	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")

	if req.BuildingID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Building ID is required"})
		return
	}

	if req.StartDate == "" || req.EndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start date and end date are required"})
		return
	}

	report, err := h.service.GetCustomerVendorReport(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/transaction-details
func (h *ReportsHandler) GetTransactionDetails(c *gin.Context) {
	var req TransactionDetailsRequest

	// Get building ID from route parameter (for building-scoped routes)
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		// Try query parameter (for legacy routes)
		buildingIDStr = c.Query("building_id")
	}
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
		}
	}

	// Parse account IDs from query (comma-separated)
	if accountIDsStr := c.Query("account_ids"); accountIDsStr != "" {
		accountIDStrs := strings.Split(accountIDsStr, ",")
		for _, idStr := range accountIDStrs {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				req.AccountIDs = append(req.AccountIDs, id)
			}
		}
	}

	if transactionType := c.Query("transaction_type"); transactionType != "" {
		req.TransactionType = &transactionType
	}

	req.StartDate = c.Query("start_date")
	req.EndDate = c.Query("end_date")

	if req.BuildingID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Building ID is required"})
		return
	}

	if req.StartDate == "" || req.EndDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start date and end date are required"})
		return
	}

	report, err := h.service.GetTransactionDetailsByAccount(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
