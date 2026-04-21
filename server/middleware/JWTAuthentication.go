package middleware

import (
	"fmt"
	"template-go/util/config"

	"github.com/golang-jwt/jwt"
)

// jwt service
type JWTService interface {
	ValidateToken(token string) (*jwt.Token, error)
}

type jwtServices struct {
	secretKey string
	issuer    string
}

// auth-jwt
func JWTAuthService(config config.Config) JWTService {
	return &jwtServices{
		secretKey: getSecretKey(config),
		issuer:    getIssuer(config),
	}
}

func getSecretKey(config config.Config) (secret string) {
	secret = config.JwtSecret
	return
}

func getIssuer(config config.Config) (issuer string) {
	issuer = config.JwtIssuer
	return
}

func (service *jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token %s", token.Header["alg"])

		}
		return []byte(service.secretKey), nil
	})

}
