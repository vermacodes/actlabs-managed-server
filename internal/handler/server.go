package handler

import (
	"actlabs-managed-server/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

type serverHandler struct {
	serverService entity.ServerService
}

func NewServerHandler(r *gin.RouterGroup, serverService entity.ServerService) {
	handler := &serverHandler{
		serverService: serverService,
	}

	r.GET("/server", handler.GetServer)
	r.PUT("/server", handler.DeployServer)
	r.DELETE("/server", handler.DestroyServer)

	r.PUT("/server/activity/:userPrincipalName", handler.UpdateActivityStatus)
}

func (h *serverHandler) GetServer(c *gin.Context) {
	server := entity.Server{}
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := h.serverService.GetServer(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, server)
}

func (h *serverHandler) DeployServer(c *gin.Context) {
	server := entity.Server{}
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := h.serverService.DeployServer(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, server)
}

func (h *serverHandler) DestroyServer(c *gin.Context) {
	server := entity.Server{}
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.serverService.DestroyServer(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func (h *serverHandler) UpdateActivityStatus(c *gin.Context) {
	userPrincipalName := c.Param("userPrincipalName")

	if err := h.serverService.UpdateActivityStatus(userPrincipalName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
