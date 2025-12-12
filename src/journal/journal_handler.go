package journal

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type JournalHandler struct {
	service *JournalService
}

func NewJournalHandler(service *JournalService) *JournalHandler {
	return &JournalHandler{service: service}
}

// POST /journals/preview or /buildings/:id/journals/preview
func (h *JournalHandler) PreviewJournal(c *gin.Context) {
	var req CreateJournalRequest
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

	preview, err := h.service.PreviewJournal(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// POST /journals or /buildings/:id/journals
func (h *JournalHandler) CreateJournal(c *gin.Context) {
	var req CreateJournalRequest
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

	response, err := h.service.CreateJournal(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PUT /journals/:id or /buildings/:id/journals/:journalId
func (h *JournalHandler) UpdateJournal(c *gin.Context) {
	var req UpdateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	journalIDStr := c.Param("journalId")
	id, err := strconv.Atoi(journalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Journal ID"})
		return
	}
	req.ID = id

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

	response, err := h.service.UpdateJournal(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /journals or /buildings/:id/journals
func (h *JournalHandler) GetJournals(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Building ID"})
		return
	}

	journals, err := h.service.GetJournals(buildingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, journals)
}

// GET /journals/:id or /buildings/:id/journals/:journalId
func (h *JournalHandler) GetJournal(c *gin.Context) {
	journalIDStr := c.Param("journalId")
	if journalIDStr == "" {
		journalIDStr = c.Param("id")
	}
	if journalIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Journal ID is required"})
		return
	}

	id, err := strconv.Atoi(journalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Journal ID"})
		return
	}

	journalResponse, err := h.service.GetJournalDetails(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, journalResponse)
}

