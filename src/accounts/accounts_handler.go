package accounts

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	service *AccountService
}

func NewAccountHandler(service *AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

// POST /accounts
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.CreateAccount(account)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErr})
		return
	}

	if otherErrors != nil {
		// Check if it's a duplicate error
		if otherErrors.Error() == "building cannot have duplicate account number" || 
		   otherErrors.Error() == "building cannot have duplicate account name" {
			c.JSON(http.StatusConflict, gin.H{"error": otherErrors.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": otherErrors.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /accounts
func (h *AccountHandler) GetAccounts(c *gin.Context) {
	accounts, err := h.service.GetAllAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GET /accounts/:id
func (h *AccountHandler) GetAccount(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	account, err := h.service.GetAccountByID(int(id))
	if err != nil {
		if err.Error() == "id does not exist" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// PUT /accounts/:id
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.UpdateAccount(id, account)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErr})
		return
	}

	if otherErrors != nil {
		// Check if it's a duplicate error for account name (only account name validation in update)
		if otherErrors.Error() == "building cannot have duplicate account name" {
			c.JSON(http.StatusConflict, gin.H{"error": otherErrors.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": otherErrors.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

