package halloween

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArteShow/Game-Manager/pkg/db"
	GetConfiguration "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/golang-jwt/jwt/v5"
)

var cache Cache

// JWT Middleware
func UserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		DatabasePath, err := GetConfiguration.GetJWTDatabasePath()
		if err != nil {
			http.Error(w, "Failed to get JWT Database path", http.StatusInternalServerError)
			return
		}

		database, err := db.OpenDataBase(DatabasePath)
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}

		jwtKey, err := db.GetCurrentJWTKey(database)
		if err != nil {
			http.Error(w, "Failed to get the JWT key", http.StatusInternalServerError)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Provide your secret key here
			return []byte(jwtKey), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}
		userID, ok := claims["userID"].(string)
		if !ok {
			http.Error(w, "userID not found in token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CreateHalloweenGame(w http.ResponseWriter, r *http.Request) {
	/*
		//Get user id
		userId := r.Context().Value("userID").(int64)

		//Get Halloween Game Name
		type CreateRequest struct{
			Name string `json:"name"`
		}
		var req CreateRequest

		defer r.Body.Close()

		err := json.Unmarshal(r.Body, &req)
		cache.AddHalloweenGame(Client{Id: int64(userId)}, )*/
}

// Start http Server
func StartTournamentHttp() error {
	port, err := GetConfiguration.GetTournamentPort()
	if err != nil {
		return err
	}
	strport := strconv.Itoa(port)

	return http.ListenAndServe(":"+strport, nil)
}
