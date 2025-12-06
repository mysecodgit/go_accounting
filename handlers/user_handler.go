package handlers

// import (
// 	"net/http"
// 	"strconv"

// 	"database/sql"

// 	"github.com/mysecodgit/go_accounting/config"

// 	"github.com/gin-gonic/gin"
// )

// // Get all users
// func GetUsers(c *gin.Context) {
// 	rows, err := config.DB.Query("SELECT id, name,username, phone FROM users")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	users := []models.UserResponse{}
// 	for rows.Next() {
// 		var user models.User
// 		err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Phone)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		users = append(users, user.ToUserReponse())
// 	}

// 	c.JSON(http.StatusOK, users)
// }

// // Get user by ID
// func GetUser(c *gin.Context) {
// 	id := c.Param("id")
// 	var user models.User
// 	err := config.DB.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).
// 		Scan(&user.ID, &user.Name, &user.Phone)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, user)
// }

// // Create user
// func CreateUser(c *gin.Context) {
// 	var user models.User

// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if err := user.Validate(); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"errors": err})
// 		return
// 	}

// 	result, err := config.DB.Exec("INSERT INTO users (name, username,phone,password) VALUES (?, ?,?,?)", user.Name, user.Username, user.Phone, user.Password)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	id, _ := result.LastInsertId()
// 	user.ID = int(id)

// 	userDto := user.ToUserReponse()

	
// 	c.JSON(http.StatusCreated, userDto)
// }

// // Update user
// func UpdateUser(c *gin.Context) {
// 	id := c.Param("id")
// 	var user models.UserUpdateRequest
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if err := user.Validate(); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"errors": err})
// 		return
// 	}

// 	_, err := config.DB.Exec("UPDATE users SET name=?, username=?,phone=? WHERE id=?", user.Name, user.Username, user.Phone, id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	user.ID, _ = strconv.Atoi(id)
// 	c.JSON(http.StatusOK, user.ToUserReponse())
// }

// // Delete user
// // func DeleteUser(c *gin.Context) {
// // 	id := c.Param("id")
// // 	_, err := config.DB.Exec("DELETE FROM users WHERE id=?", id)
// // 	if err != nil {
// // 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// // 		return
// // 	}

// // 	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
// // }
