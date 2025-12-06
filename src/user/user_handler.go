package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

// POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user RegisterUserRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	response, validationErr, otherErrors := h.service.Register(user)

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

// GET /users
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	stringId := c.Param("id")
	id, err := strconv.Atoi(stringId)
	
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	users, err := h.service.GetUserByID(int(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
