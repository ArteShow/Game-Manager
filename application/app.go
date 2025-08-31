package application

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/ArteShow/Game-Manager/pkg/registartion"
	"github.com/golang-jwt/jwt/v4"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		err, secret := registartion.GetJWTKey()
		if err != nil {
			http.Error(w, "Error loading JWT key", http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok {
			http.Error(w, "Invalid userID in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RegistereNewUser(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post("http://localhost:8081/internal/register", "application/json", r.Body)
	if err != nil {
		http.Error(w, "Error while creating new user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
}

func LoginNewUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Login of new user")

	resp, err := http.Post("http://localhost:8081/internal/login", "application/json", r.Body)
	if err != nil {
		http.Error(w, "Error while creating new user", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read the response body", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func StartApplicationServer() error {
	log.Println("Starting Application Server")
	port, err := getconfig.GetApplicationPort()
	if err != nil {
		log.Fatal(err)
	}
	portStr := ":" + strconv.Itoa(port)

	http.HandleFunc("/reg", RegistereNewUser)
	http.HandleFunc("/login", LoginNewUser)

	return http.ListenAndServe(portStr, nil)
}
