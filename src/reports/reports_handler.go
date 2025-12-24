package reports

import (
	"net/http"
	"strconv"

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

// GET /reports/trial-balance
func (h *ReportsHandler) GetTrialBalance(c *gin.Context) {
	var req TrialBalanceRequest

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

	report, err := h.service.GetTrialBalance(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/transaction-details-by-account
func (h *ReportsHandler) GetTransactionDetailsByAccount(c *gin.Context) {
	var req TransactionDetailsByAccountRequest

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

	// Support multiple account_id query parameters
	accountIDStrs := c.QueryArray("account_id")
	if len(accountIDStrs) > 0 {
		req.AccountIDs = []int{}
		for _, accountIDStr := range accountIDStrs {
			accountID, err := strconv.Atoi(accountIDStr)
			if err == nil && accountID > 0 {
				req.AccountIDs = append(req.AccountIDs, accountID)
			}
		}
	}

	if unitIDStr := c.Query("unit_id"); unitIDStr != "" {
		unitID, err := strconv.Atoi(unitIDStr)
		if err == nil {
			req.UnitID = &unitID
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

	report, err := h.service.GetTransactionDetailsByAccount(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/customer-balance-summary
func (h *ReportsHandler) GetCustomerBalanceSummary(c *gin.Context) {
	var req CustomerBalanceSummaryRequest

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

	report, err := h.service.GetCustomerBalanceSummary(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/customer-balance-details
func (h *ReportsHandler) GetCustomerBalanceDetails(c *gin.Context) {
	var req CustomerBalanceDetailsRequest

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

	// Get people_id from query (optional)
	if peopleIDStr := c.Query("people_id"); peopleIDStr != "" {
		peopleID, err := strconv.Atoi(peopleIDStr)
		if err == nil {
			req.PeopleID = &peopleID
		}
	}

	if req.BuildingID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Building ID is required"})
		return
	}

	report, err := h.service.GetCustomerBalanceDetails(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/profit-and-loss-standard
func (h *ReportsHandler) GetProfitAndLossStandard(c *gin.Context) {
	var req ProfitAndLossStandardRequest

	// Get building ID from route parameter
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		buildingIDStr = c.Query("building_id")
	}
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
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

	report, err := h.service.GetProfitAndLossStandard(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GET /reports/profit-and-loss-by-unit
func (h *ReportsHandler) GetProfitAndLossByUnit(c *gin.Context) {
	var req ProfitAndLossByUnitRequest

	// Get building ID from route parameter
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		buildingIDStr = c.Query("building_id")
	}
	if buildingIDStr != "" {
		buildingID, err := strconv.Atoi(buildingIDStr)
		if err == nil {
			req.BuildingID = buildingID
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

	report, err := h.service.GetProfitAndLossByUnit(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}
