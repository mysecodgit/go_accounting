package main

import (
    "github.com/gin-gonic/gin"
    "github.com/mysecodgit/go_accounting/config"
    "github.com/mysecodgit/go_accounting/routes"
 
)

func main() {
    r := gin.Default()

    config.ConnectDatabase()
    routes.SetupRoutes(r)

	


    r.Run(":8083")
}
