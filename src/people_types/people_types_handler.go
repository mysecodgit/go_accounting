package people_types

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PeopleTypeHandler struct {
	service *PeopleTypeService
}

func NewPeopleTypeHandler(service *PeopleTypeService) *PeopleTypeHandler {
	return &PeopleTypeHandler{service: service}
}

// POST /people-types
func (h *PeopleTypeHandler) CreatePeopleType(c *gin.Context) {
	var peopleType PeopleType
	if err := c.ShouldBindJSON(&peopleType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.CreatePeopleType(peopleType)

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

// GET /people-types
func (h *PeopleTypeHandler) GetPeopleTypes(c *gin.Context) {
	peopleTypes, err := h.service.GetAllPeopleTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, peopleTypes)
}

// GET /people-types/:id
func (h *PeopleTypeHandler) GetPeopleType(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	peopleType, err := h.service.GetPeopleTypeByID(int(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, peopleType)
}

// PUT /people-types/:id
func (h *PeopleTypeHandler) UpdatePeopleType(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var updateReq UpdatePeopleTypeRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.UpdatePeopleType(id, updateReq)

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

