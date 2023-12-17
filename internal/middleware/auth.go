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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate"
	"golang.org/x/exp/slog"
)

func Auth(rateLimiter *redis_rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Debug("Auth Middleware")
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			slog.Error("no auth token provided")
			allow := handleBadRequest(c, rateLimiter)
			if allow {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			return
		}

		err := handleAccessToken(c, accessToken)
		if err != nil {
			handleBadRequest(c, rateLimiter)
			return
		}
		c.Next()
	}
}

func handleAccessToken(c *gin.Context, accessToken string) error {
	body, _ := io.ReadAll(c.Request.Body)
	server := entity.Server{}
	if err := json.Unmarshal(body, &server); err != nil {
		slog.Error("error binding json", slog.String("error", err.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}

	// Reassign the body so it can be read again in the handler
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	splitToken := strings.Split(accessToken, "Bearer ")
	if len(splitToken) < 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return errors.New("found something in the Authorization header, but it's not a bearer token")
	}

	ok, err := helper.VerifyToken(accessToken, server.UserPrincipalId)
	if err != nil || !ok {
		slog.Error("token verification failed", slog.String("error", err.Error()))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return err
	}

	return nil
}

func handleBadRequest(c *gin.Context, rateLimiter *redis_rate.Limiter) bool {
	ip := c.ClientIP()

	count, delay, allow := rateLimiter.Allow(ip, 10, time.Minute*10)

	slog.Info("bad request",
		slog.String("ip", ip),
		slog.Int64("count", count),
		slog.Duration("delay", delay),
		slog.Bool("allow", allow),
	)

	if !allow {
		slog.Error("too many bad requests, ip blocked",
			slog.String("ip", ip),
		)

		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many bad requests, try again later"})
	}

	return allow
}
