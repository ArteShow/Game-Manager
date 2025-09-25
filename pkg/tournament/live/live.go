package live

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ArteShow/Game-Manager/models"
	"github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

func NewServer() *models.LiveServer {
	return &models.LiveServer{
		Tournaments: []models.Tournament{},
		Mu:          sync.Mutex{},
		Broadcast:   make(chan models.BroadcastMessage),
		Tournament:  make(chan models.TournamentMessage),
		Clients:     []models.Client{},
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type contextKey string

const UserIDKey contextKey = "userID"

func JWTMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			DatabasePath, err := getconfig.GetJWTDatabasePath()
			if err != nil {
				http.Error(w, "Failed to get the Database path", http.StatusInternalServerError)
				return
			}

			Database, err := db.OpenDataBase(DatabasePath)
			if err != nil {
				http.Error(w, "Failed to open the database", http.StatusInternalServerError)
				return
			}

			jwtSecret, err := db.GetCurrentJWTKey(Database)
			if err != nil {
				http.Error(w, "Failed to get jwt key", http.StatusInternalServerError)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]

			type claimsT struct {
				UserID string `json:"userID"`
				jwt.RegisteredClaims
			}

			token, err := jwt.ParseWithClaims(tokenStr, &claimsT{}, func(t *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*claimsT)
			if !ok || claims.UserID == "" {
				http.Error(w, "userID not found in token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func StartLive() error {
	LiveServer := NewServer()

	portInt, err := getconfig.GetTournamentPort()
	if err != nil {
		log.Fatal(err)
	}
	strPort := strconv.Itoa(portInt)

	http.Handle("/add", JWTMiddleware()(LiveServer.AddTournament()))
	http.Handle("/delet", JWTMiddleware()(LiveServer.DeleteTournament()))

	ws := &WsServer{Server: *LiveServer}
	http.Handle("/ws", JWTMiddleware()(ws.StartWs()))
	return http.ListenAndServe(":"+strPort, nil)
}
