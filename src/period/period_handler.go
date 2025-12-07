package period

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PeriodHandler struct {
	service *PeriodService
}

func NewPeriodHandler(service *PeriodService) *PeriodHandler {
	return &PeriodHandler{service: service}
}

// POST /periods
func (h *PeriodHandler) CreatePeriod(c *gin.Context) {
	var period Period
	if err := c.ShouldBindJSON(&period); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.CreatePeriod(period)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErr})
		return
	}

	if otherErrors != nil {
		// Check if it's a duplicate period error
		if otherErrors.Error() == "building cannot have duplicate period with the same start and end date" {
			c.JSON(http.StatusConflict, gin.H{"error": otherErrors.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": otherErrors.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /periods
func (h *PeriodHandler) GetPeriods(c *gin.Context) {
	periods, err := h.service.GetAllPeriods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, periods)
}

// GET /periods/:id
func (h *PeriodHandler) GetPeriod(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	period, err := h.service.GetPeriodByID(int(id))
	if err != nil {
		if err.Error() == "id does not exist" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, period)
}

// PUT /periods/:id
func (h *PeriodHandler) UpdatePeriod(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var period Period
	if err := c.ShouldBindJSON(&period); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.UpdatePeriod(id, period)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErr})
		return
	}

	if otherErrors != nil {
		// Check if it's a duplicate period error
		if otherErrors.Error() == "building cannot have duplicate period with the same start and end date" {
			c.JSON(http.StatusConflict, gin.H{"error": otherErrors.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": otherErrors.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

