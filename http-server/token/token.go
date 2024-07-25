package token

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
)

var tokenSalt = os.Getenv("TOKEN_SALT")

type AuthClaims struct {
	jwt.RegisteredClaims
	TokenSum int
}

// generate an access token. Due to the fact that we have only 1 user we will have only one access rights
func GenerateAccessToken(claims AuthClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSalt))
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return signedToken, nil
}
