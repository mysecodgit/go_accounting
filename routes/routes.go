package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mysecodgit/go_accounting/config"
	_ "github.com/mysecodgit/go_accounting/handlers"
	"github.com/mysecodgit/go_accounting/src/user"
)

func SetupRoutes(r *gin.Engine) {

	userRepo := user.NewUserRepository(config.DB)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	userRoutes := r.Group("/api/users")
	{
		userRoutes.GET("", userHandler.GetUsers)
		userRoutes.GET("/:id", userHandler.GetUser)
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.PUT("/:id", userHandler.UpdateUser)
		// userRoutes.DELETE("/:id", handlers.DeleteUser)
	}
}
