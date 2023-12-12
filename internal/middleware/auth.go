package middleware

import (
	"actlabs-managed-server/internal/entity"
	"actlabs-managed-server/internal/helper"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Debug("Auth Middleware")
		accessToken := c.GetHeader("Authorization")
		err := handleAccessToken(c, accessToken)
		if err != nil {
			return
		}
		c.Next()
	}
}

func handleAccessToken(c *gin.Context, accessToken string) error {
	body, _ := io.ReadAll(c.Request.Body)
	server := entity.Server{}
	if err := json.Unmarshal(body, &server); err != nil {
		slog.Error("Error binding JSON", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	// Reassign the body so it can be read again in the handler
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	splitToken := strings.Split(accessToken, "Bearer ")
	if len(splitToken) < 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no auth token provided"})
		return errors.New("no auth token provided")
	}

	ok, err := helper.VerifyToken(accessToken, server.UserPrincipalId)
	if err != nil || !ok {
		slog.Error("Token is not issued by AAD", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return err
	}

	return nil
}
