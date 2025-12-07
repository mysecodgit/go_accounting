package account_types

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountTypeHandler struct {
	service *AccountTypeService
}

func NewAccountTypeHandler(service *AccountTypeService) *AccountTypeHandler {
	return &AccountTypeHandler{service: service}
}

// POST /account-types
func (h *AccountTypeHandler) CreateAccountType(c *gin.Context) {
	var accountType AccountType
	if err := c.ShouldBindJSON(&accountType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.CreateAccountType(accountType)

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

// GET /account-types
func (h *AccountTypeHandler) GetAccountTypes(c *gin.Context) {
	accountTypes, err := h.service.GetAllAccountTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accountTypes)
}

// GET /account-types/:id
func (h *AccountTypeHandler) GetAccountType(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	accountType, err := h.service.GetAccountTypeByID(int(id))
	if err != nil {
		if err.Error() == "id does not exist" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accountType)
}

// PUT /account-types/:id
func (h *AccountTypeHandler) UpdateAccountType(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var accountType AccountType
	if err := c.ShouldBindJSON(&accountType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.UpdateAccountType(id, accountType)

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

