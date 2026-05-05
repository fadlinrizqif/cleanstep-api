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
