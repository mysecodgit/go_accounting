package unit

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UnitHandler struct {
	service *UnitService
}

func NewUnitHandler(service *UnitService) *UnitHandler {
	return &UnitHandler{service: service}
}

// POST /units
func (h *UnitHandler) CreateUnit(c *gin.Context) {
	var unit Unit
	if err := c.ShouldBindJSON(&unit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.CreateUnit(unit)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErr})
		return
	}

	if otherErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": otherErrors.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /units
func (h *UnitHandler) GetUnits(c *gin.Context) {
	units, err := h.service.GetAllUnits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, units)
}

// GET /buildings/:id/units
func (h *UnitHandler) GetUnitsByBuilding(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Building ID"})
		return
	}

	units, err := h.service.GetUnitsByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, units)
}

// GET /units/:id or /buildings/:id/units/:unitId
func (h *UnitHandler) GetUnit(c *gin.Context) {
	// Check for unitId first (building-scoped route), then fall back to id
	stringId := c.Param("unitId")
	if stringId == "" {
		stringId = c.Param("id")
	}
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	unit, err := h.service.GetUnitByID(int(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, unit)
}

// PUT /units/:id or /buildings/:id/units/:unitId
func (h *UnitHandler) UpdateUnit(c *gin.Context) {
	// Check for unitId first (building-scoped route), then fall back to id
	stringId := c.Param("unitId")
	if stringId == "" {
		stringId = c.Param("id")
	}
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var unit Unit
	if err := c.ShouldBindJSON(&unit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.UpdateUnit(id, unit)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErr})
		return
	}

	if otherErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": otherErrors.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

