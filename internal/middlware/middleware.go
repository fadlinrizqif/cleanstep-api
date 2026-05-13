package middlware

import (
	"log"
	"net/http"
	"time"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(app *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		getToken, err := c.Cookie("access_token")

		if err == nil || getToken != "" {
			userID, err := auth.ValidateJWT(getToken, app.SeverSecret)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.Set("getUserID", userID)
			c.Next()
			return
		}

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil || refreshToken == "" {
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
		}

		userID, err := auth.ValidateRefreshToken(refreshToken, app.DBqueries, c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			log.Println("error disini")
			return
		}

		jwtDuration := time.Duration(60) * time.Minute
		newToken, err := auth.MakeJWT(userID, app.SeverSecret, jwtDuration)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		header := c.Request.Header
		getForwardedProto := header.Get("X-Forwarded-Proto")
		secure := c.Request.TLS != nil || getForwardedProto == "https"
		c.SetCookieData(&http.Cookie{
			Name:     "access_token",
			Value:    newToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   60 * 60,
		})

		c.Set("getUserID", userID)
		c.Next()

	}
}

func AuthMidtrans() gin.HandlerFunc {
	return func(c *gin.Context) {
		var midtransKey struct {
			SignatureKey string `json:"signature_key"`
		}

		if err := c.BindJSON(&midtransKey); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong request"})
			return
		}

		if midtransKey.SignatureKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()

	}
}
