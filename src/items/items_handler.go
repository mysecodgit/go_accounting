package items

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ItemHandler struct {
	service *ItemService
}

func NewItemHandler(service *ItemService) *ItemHandler {
	return &ItemHandler{service: service}
}

// POST /items
func (h *ItemHandler) CreateItem(c *gin.Context) {
	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.CreateItem(item)

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

// GET /items
func (h *ItemHandler) GetItems(c *gin.Context) {
	items, err := h.service.GetAllItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GET /buildings/:id/items
func (h *ItemHandler) GetItemsByBuilding(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Building ID"})
		return
	}

	items, err := h.service.GetItemsByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GET /items/:id or /buildings/:id/items/:itemId
func (h *ItemHandler) GetItem(c *gin.Context) {
	// Check for itemId first (building-scoped route), then fall back to id
	stringId := c.Param("itemId")
	if stringId == "" {
		stringId = c.Param("id")
	}
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	item, err := h.service.GetItemByID(int(id))
	if err != nil {
		if err.Error() == "id does not exist" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// PUT /items/:id or /buildings/:id/items/:itemId
func (h *ItemHandler) UpdateItem(c *gin.Context) {
	// Check for itemId first (building-scoped route), then fall back to id
	stringId := c.Param("itemId")
	if stringId == "" {
		stringId = c.Param("id")
	}
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.UpdateItem(id, item)

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

