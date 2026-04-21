package middleware

import (
	"net/http"
	"template-go/util/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthorizeJWT(config config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		const bearerSchema = "Bearer "
		var authHeader string
		cookie, err := c.Request.Cookie("Authorization")
		if err == nil {
			authHeader = cookie.Value
		}
		if len(authHeader) < 1 {
			authHeader = c.GetHeader("Authorization")
		}
		if len(authHeader) < len(bearerSchema) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": http.StatusUnauthorized, "message": "Unauthorized access, Invalid/Missing token"})
			return
		}

		tokenString := authHeader[len(bearerSchema):]
		token, err := JWTAuthService(config).ValidateToken(tokenString)
		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			// save user's name, role and id
			c.Set("ID", int(claims["ID"].(float64)))
			c.Set("fullName", claims["FullName"])
			c.Set("email", claims["Email"])
			c.Set("role", claims["Role"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": http.StatusUnauthorized, "message": "Unauthorized access, Invalid token"})
		}

	}
}
