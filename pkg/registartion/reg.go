package registartion

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/ArteShow/Game-Manager/models"
	"github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/golang-jwt/jwt/v5"
)

func GetJWTKey() (error, string) {
	path, err := getconfig.GetJWTDatabasePath()
	if err != nil {
		log.Fatal(err)
		return err, ""
	}

	DB, err := db.OpenDataBase(path)
	if err != nil {
		log.Fatal(err)
		return err, ""
	}

	key, err3 := db.GetCurrentJWTKey(DB)
	if err3 != nil {
		log.Fatal(err3)
		return err3, ""
	}

	return nil, key
}

func CheckUserCredentials(username, password string, Databse *sql.DB) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = ? AND password = ?`

	var count int
	err := Databse.QueryRow(query, username, password).Scan(&count)
	if err != nil {
		log.Println("Error checking user credentials:", err)
		return false, err
	}

	if count == 0 {
		return false, errors.New("invalid username or password")
	}

	return true, nil
}

func CheckUserData(userData models.Login) (bool, error) {
	path, err := getconfig.GetUserdatabasePath()
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	Database, err := db.OpenDataBase(path)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	ok, err := CheckUserCredentials(userData.Username, userData.Passwword, Database)
	if ok {
		return true, nil
	} else {
		return false, err
	}

}

func GenerateJWT(userID int64) (string, error) {
	log.Println("Generating a new key")
	claims := jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	err, jwtKey := GetJWTKey()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return token.SignedString([]byte(jwtKey))
}
