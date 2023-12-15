package main

import (
	"actlabs-managed-server/internal/auth"
	"actlabs-managed-server/internal/config"
	"actlabs-managed-server/internal/handler"
	"actlabs-managed-server/internal/logger"
	"actlabs-managed-server/internal/middleware"
	"actlabs-managed-server/internal/repository"
	"actlabs-managed-server/internal/service"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	logger.SetupLogger()
	appConfig := config.NewConfig()
	auth := auth.NewAuth(appConfig)

	serverRepository, err := repository.NewServerRepository(appConfig, auth)
	if err != nil {
		panic(err)
	}

	serverService := service.NewServerService(serverRepository, appConfig)

	router := gin.Default()
	router.SetTrustedProxies(nil)

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173", "https://ashisverma.z13.web.core.windows.net", "https://actlabs.z13.web.core.windows.net", "https://actlabsbeta.z13.web.core.windows.net", "https://actlabs.azureedge.net", "https://*.azurewebsites.net"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Authorization", "Content-Type"}

	router.Use(cors.New(config))
	router.Use(middleware.Auth())

	handler.NewServerHandler(router.Group("/"), serverService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8883"
	}
	router.Run(":" + port)
}
