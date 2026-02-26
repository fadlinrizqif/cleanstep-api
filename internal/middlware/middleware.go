package middlware

import (
	"net/http"

	"github.com/fadlinrizqif/cleanstep-api/internal/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(sercretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		getToken, err := c.Cookie("access_token")

		if err != nil || getToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		userID, err := auth.ValidateJWT(getToken, sercretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("getUserID", userID)

		c.Next()

	}
}
