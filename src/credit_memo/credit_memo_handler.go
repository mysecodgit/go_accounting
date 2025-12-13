package credit_memo

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreditMemoHandler struct {
	service *CreditMemoService
}

func NewCreditMemoHandler(service *CreditMemoService) *CreditMemoHandler {
	return &CreditMemoHandler{service: service}
}

// POST /credit-memos/preview or /buildings/:id/credit-memos/preview
func (h *CreditMemoHandler) PreviewCreditMemo(c *gin.Context) {
	var req CreateCreditMemoRequest
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

	preview, err := h.service.PreviewCreditMemo(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// POST /credit-memos or /buildings/:id/credit-memos
func (h *CreditMemoHandler) CreateCreditMemo(c *gin.Context) {
	var req CreateCreditMemoRequest
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

	response, err := h.service.CreateCreditMemo(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PUT /credit-memos/:id or /buildings/:id/credit-memos/:creditMemoId
func (h *CreditMemoHandler) UpdateCreditMemo(c *gin.Context) {
	var req UpdateCreditMemoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Get credit memo ID from route
	creditMemoIDStr := c.Param("creditMemoId")
	if creditMemoIDStr == "" {
		creditMemoIDStr = c.Param("id")
	}
	if creditMemoIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credit memo ID is required"})
		return
	}

	creditMemoID, err := strconv.Atoi(creditMemoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credit memo ID"})
		return
	}
	req.ID = creditMemoID

	// Get building_id from route if building-scoped
	buildingIDStr := c.Param("id")
	if buildingIDStr != "" && buildingIDStr != creditMemoIDStr {
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

	response, err := h.service.UpdateCreditMemo(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /credit-memos/:id or /buildings/:id/credit-memos/:creditMemoId
func (h *CreditMemoHandler) GetCreditMemoByID(c *gin.Context) {
	// Get credit memo ID from route
	creditMemoIDStr := c.Param("creditMemoId")
	if creditMemoIDStr == "" {
		creditMemoIDStr = c.Param("id")
	}
	if creditMemoIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Credit memo ID is required"})
		return
	}

	creditMemoID, err := strconv.Atoi(creditMemoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credit memo ID"})
		return
	}

	response, err := h.service.GetCreditMemoByID(creditMemoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /credit-memos or /buildings/:id/credit-memos
func (h *CreditMemoHandler) GetCreditMemosByBuildingID(c *gin.Context) {
	// Get building ID from route parameter (for building-scoped routes)
	buildingIDStr := c.Param("id")
	if buildingIDStr == "" {
		buildingIDStr = c.Query("building_id")
	}
	if buildingIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Building ID is required"})
		return
	}

	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	creditMemos, err := h.service.GetCreditMemosByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, creditMemos)
}

