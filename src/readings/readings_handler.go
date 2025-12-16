package readings

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReadingHandler struct {
	service *ReadingService
}

func NewReadingHandler(service *ReadingService) *ReadingHandler {
	return &ReadingHandler{service: service}
}

// GET /buildings/:id/readings
func (h *ReadingHandler) GetReadings(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	// Get optional status filter
	status := c.Query("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	readings, err := h.service.GetReadingsByBuildingID(buildingID, statusPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, readings)
}

// POST /buildings/:id/readings
func (h *ReadingHandler) CreateReading(c *gin.Context) {
	var req CreateReadingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, err := h.service.CreateReading(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /buildings/:id/readings/:readingId
func (h *ReadingHandler) GetReadingByID(c *gin.Context) {
	readingIDStr := c.Param("readingId")
	readingID, err := strconv.Atoi(readingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reading ID"})
		return
	}

	response, err := h.service.GetReadingByID(readingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PUT /buildings/:id/readings/:readingId
func (h *ReadingHandler) UpdateReading(c *gin.Context) {
	var req UpdateReadingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	readingIDStr := c.Param("readingId")
	readingID, err := strconv.Atoi(readingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reading ID"})
		return
	}
	req.ID = readingID

	response, err := h.service.UpdateReading(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /buildings/:id/readings/unit/:unitId
func (h *ReadingHandler) GetReadingsByUnitID(c *gin.Context) {
	unitIDStr := c.Param("unitId")
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	readings, err := h.service.GetReadingsByUnitID(unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, readings)
}

// DELETE /buildings/:id/readings/:readingId
func (h *ReadingHandler) DeleteReading(c *gin.Context) {
	readingIDStr := c.Param("readingId")
	readingID, err := strconv.Atoi(readingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reading ID"})
		return
	}

	err = h.service.DeleteReading(readingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reading deleted successfully"})
}

// POST /buildings/:id/readings/import
func (h *ReadingHandler) BulkImportReadings(c *gin.Context) {
	var req BulkImportReadingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, err := h.service.BulkImportReadings(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  err.Error(),
			"result": response,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /buildings/:id/readings/latest
func (h *ReadingHandler) GetLatestReading(c *gin.Context) {
	itemIDStr := c.Query("item_id")
	unitIDStr := c.Query("unit_id")

	if itemIDStr == "" || unitIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id and unit_id are required"})
		return
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	reading, err := h.service.GetLatestReadingByItemAndUnit(itemID, unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if reading == nil {
		c.JSON(http.StatusOK, gin.H{"reading": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reading": reading})
}
