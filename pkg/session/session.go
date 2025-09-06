package session

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/ArteShow/Game-Manager/models"
	"github.com/ArteShow/Game-Manager/pkg/profiles"
	"github.com/ArteShow/Game-Manager/pkg/rooms"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type SessionServer struct {
	HubCache    *models.HubCache
	RoomServers map[int64]*rooms.RoomServer
	Database    *sql.DB
	JWTSecret   []byte
}

func NewSessionServer(hubCache *models.HubCache, roomServers map[int64]*rooms.RoomServer, database *sql.DB, secret []byte) *SessionServer {
	return &SessionServer{
		HubCache:    hubCache,
		RoomServers: roomServers,
		Database:    database,
		JWTSecret:   secret,
	}
}

func (s *SessionServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleWS: New connection attempt")
	roomIDStr := r.URL.Query().Get("roomID")
	profileIDStr := r.URL.Query().Get("profileID")
	if roomIDStr == "" || profileIDStr == "" {
		http.Error(w, "roomID and profileID required", http.StatusBadRequest)
		log.Println("HandleWS: Missing roomID or profileID")
		return
	}

	roomID, _ := strconv.ParseInt(roomIDStr, 10, 64)
	profileID, _ := strconv.ParseInt(profileIDStr, 10, 64)
	log.Printf("HandleWS: roomID=%d profileID=%d\n", roomID, profileID)

	authHeader := r.Header.Get("Authorization")
	var tokenStr string
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr = authHeader[7:]
	} else {
		tokenStr = r.URL.Query().Get("token")
	}

	if tokenStr == "" {
		http.Error(w, "Authorization header or token query param required", http.StatusUnauthorized)
		log.Println("HandleWS: Missing or invalid Authorization header and no token query param")
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return s.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		log.Println("HandleWS: JWT parse/validation failed:", err)
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	userID := int64(claims["userID"].(float64))
	log.Printf("HandleWS: userID=%d authenticated\n", userID)

	profile, err := profiles.GetProfileByID(s.Database, userID, profileID)
	if err != nil {
		http.Error(w, "Profile not found", http.StatusBadRequest)
		log.Println("HandleWS: Profile not found for userID", userID)
		return
	}

	s.HubCache.Mu.Lock()
	s.HubCache.UserProfiles[userID] = profile
	s.HubCache.Mu.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("HandleWS: WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()
	log.Printf("HandleWS: WebSocket upgraded for user %d\n", userID)

	roomSrv, ok := s.RoomServers[roomID]
	if !ok {
		log.Printf("HandleWS: Room %d does not exist\n", roomID)
		return
	}

	roomSrv.AddConnection(userID, conn)
	defer roomSrv.RemoveConnection(userID)

	roomSrv.Join <- userID
	defer func() { roomSrv.Leave <- userID }()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("HandleWS: User %d disconnected from room %d\n", userID, roomID)
			break
		}
		var roomMsg models.RoomMessage
		if err := json.Unmarshal(msg, &roomMsg); err != nil {
			log.Println("HandleWS: Invalid message format:", err)
			continue
		}
		roomMsg.UserID = userID
		roomSrv.Messages <- roomMsg
		log.Printf("HandleWS: Forwarded message from user %d: %s\n", userID, roomMsg.Message)
	}
}
