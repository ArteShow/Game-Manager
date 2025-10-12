package halloween

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ArteShow/Game-Manager/pkg/db"
	GetConfiguration "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
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

// Create new hw server
func CreateHalloweenGame(w http.ResponseWriter, r *http.Request) {
	//Get user id
	userId := r.Context().Value("userID").(int64)

	//Get Halloween Game Name
	type CreateRequest struct {
		Name string `json:"name"`
	}
	var req CreateRequest

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read the body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Failed to unmarshal the body", http.StatusInternalServerError)
		return
	}

	//Return New Halloween Game Id
	id := cache.AddHalloweenGame(Client{Id: int64(userId)}, req.Name)
	bytes, err := json.Marshal(id)
	if err != nil {
		http.Error(w, "Failed to marshal to bytes", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(bytes)
}

// stop and delete hw server
func DeleteServer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	type Request struct {
		HwsID int64 `json:"id"`
	}
	var req Request

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to read the body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Failed to unmarshal the body", http.StatusInternalServerError)
		return
	}

	var connection *websocket.Conn
	for _, hw := range cache.HalloweenGame {
		for _, cl := range hw.Players {
			if cl.Id == userID {
				connection = &cl.Conn
			}
		}
	}

	for i, hws := range cache.HalloweenServers {
		if hws.Id == req.HwsID {
			stopMessage := StopMessage{
				AdminID: userID,
				Conn:    connection,
				Type:    "STOP",
			}

			hws.Stop <- stopMessage

			cache.HalloweenServers = append(cache.HalloweenServers[:i], cache.HalloweenServers[i+1:]...)
			break
		}
	}

	for i, hw := range cache.HalloweenGame {
		if hw.Id == req.HwsID {
			cache.HalloweenGame = append(cache.HalloweenGame[i:], cache.HalloweenGame[i+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Start http Server
func StartTournamentHttp() error {
	//Initialize new cache
	cache = Cache{
		Mu:            sync.Mutex{},
		HalloweenGame: []HalloweenGame{},
	}

	port, err := GetConfiguration.GetTournamentPort()
	if err != nil {
		return err
	}
	strport := strconv.Itoa(port)

	//endpoints
	http.Handle("/hw/delete", UserIDMiddleware(http.HandlerFunc(DeleteServer)))
	http.Handle("/hw/ws", UserIDMiddleware(http.HandlerFunc(HalloweenWebsocketServer)))
	http.Handle("/hw/add", UserIDMiddleware(http.HandlerFunc(CreateHalloweenGame)))
	return http.ListenAndServe(":"+strport, nil)
}
