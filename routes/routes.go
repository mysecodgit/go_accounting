package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mysecodgit/go_accounting/config"
	_ "github.com/mysecodgit/go_accounting/handlers"
	"github.com/mysecodgit/go_accounting/src/building"
	"github.com/mysecodgit/go_accounting/src/people"
	"github.com/mysecodgit/go_accounting/src/people_types"
	"github.com/mysecodgit/go_accounting/src/unit"
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

	buildingRepo := building.NewBuildingRepository(config.DB)
	buildingService := building.NewBuildingService(buildingRepo)
	buildingHandler := building.NewBuildingHandler(buildingService)

	buildingRoutes := r.Group("/api/buildings")
	{
		buildingRoutes.GET("", buildingHandler.GetBuildings)
		buildingRoutes.GET("/:id", buildingHandler.GetBuilding)
		buildingRoutes.POST("", buildingHandler.CreateBuilding)
		buildingRoutes.PUT("/:id", buildingHandler.UpdateBuilding)
	}

	unitRepo := unit.NewUnitRepository(config.DB)
	unitService := unit.NewUnitService(unitRepo)
	unitHandler := unit.NewUnitHandler(unitService)

	unitRoutes := r.Group("/api/units")
	{
		unitRoutes.GET("", unitHandler.GetUnits)
		unitRoutes.GET("/:id", unitHandler.GetUnit)
		unitRoutes.POST("", unitHandler.CreateUnit)
		unitRoutes.PUT("/:id", unitHandler.UpdateUnit)
	}

	peopleTypeRepo := people_types.NewPeopleTypeRepository(config.DB)
	peopleTypeService := people_types.NewPeopleTypeService(peopleTypeRepo)
	peopleTypeHandler := people_types.NewPeopleTypeHandler(peopleTypeService)

	peopleTypeRoutes := r.Group("/api/people-types")
	{
		peopleTypeRoutes.GET("", peopleTypeHandler.GetPeopleTypes)
		peopleTypeRoutes.GET("/:id", peopleTypeHandler.GetPeopleType)
		peopleTypeRoutes.POST("", peopleTypeHandler.CreatePeopleType)
		peopleTypeRoutes.PUT("/:id", peopleTypeHandler.UpdatePeopleType)
	}

	personRepo := people.NewPersonRepository(config.DB)
	personService := people.NewPersonService(personRepo)
	personHandler := people.NewPersonHandler(personService)

	peopleRoutes := r.Group("/api/people")
	{
		peopleRoutes.GET("", personHandler.GetPeople)
		peopleRoutes.GET("/:id", personHandler.GetPerson)
		peopleRoutes.POST("", personHandler.CreatePerson)
		peopleRoutes.PUT("/:id", personHandler.UpdatePerson)
	}
}
