package registartion

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID int64) (string, error) {
	log.Println("Generating a new key")
	claims := jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	log.Println(token) //delet
	//return token.SignedString(jwtKey)
	return "", nil
}
