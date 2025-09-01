package application

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArteShow/Game-Manager/models"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/ArteShow/Game-Manager/pkg/registartion"
	"github.com/golang-jwt/jwt/v4"
)

type CtxKey string

const UserIDKey CtxKey = "userID"

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

		userIDFloat, ok := claims["userID"].(float64)
		if !ok {
			http.Error(w, "Invalid userID in token", http.StatusUnauthorized)
			return
		}
		userID := int64(userIDFloat)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RegistereNewUser(w http.ResponseWriter, r *http.Request) {
	port, err := getconfig.GetInternalPort()
	if err != nil {
		http.Error(w, "Internal server port not available", http.StatusInternalServerError)
		return
	}
	url := "http://localhost:" + strconv.Itoa(port) + "/internal/register"

	resp, err := http.Post(url, "application/json", r.Body)
	if err != nil {
		http.Error(w, "Error while creating new user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
}

func LoginNewUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Login of new user")

	port, err := getconfig.GetInternalPort()
	if err != nil {
		http.Error(w, "Internal server port not available", http.StatusInternalServerError)
		return
	}
	url := "http://localhost:" + strconv.Itoa(port) + "/internal/login"

	resp, err := http.Post(url, "application/json", r.Body)
	if err != nil {
		http.Error(w, "Error while logging in user", http.StatusInternalServerError)
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

// Profiles
func CreataeNewProfile(w http.ResponseWriter, r *http.Request) {
	userIDCtx := r.Context().Value(UserIDKey)
	if userIDCtx == nil {
		http.Error(w, "userID not found in context", http.StatusUnauthorized)
		return
	}
	userID := userIDCtx.(int64)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var profileData models.ProfileData
	err = json.Unmarshal(body, &profileData)
	if err != nil {
		http.Error(w, "Error unmarshaling body", http.StatusInternalServerError)
		return
	}

	profileData.UserID = userID
	newBody, err := json.Marshal(profileData)
	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	port, err := getconfig.GetInternalPort()
	if err != nil {
		http.Error(w, "Internal server port not available", http.StatusInternalServerError)
		return
	}
	url := "http://localhost:" + strconv.Itoa(port) + "/internal/createProfile"

	resp, err := http.Post(url, "application/json", bytes.NewReader(newBody))
	if err != nil {
		http.Error(w, "Error sending request to internal server", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
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

	http.Handle("/createProfile", JWTMiddleware(http.HandlerFunc(CreataeNewProfile)))

	return http.ListenAndServe(portStr, nil)
}
