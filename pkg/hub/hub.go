package hub

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ArteShow/Game-Manager/models"
	"github.com/ArteShow/Game-Manager/pkg/db"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/ArteShow/Game-Manager/pkg/profiles"
	"github.com/ArteShow/Game-Manager/pkg/rooms"
	"github.com/ArteShow/Game-Manager/pkg/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Hub struct {
	Cache       *models.HubCache
	RoomServers map[int64]*rooms.RoomServer
	Mu          sync.Mutex
	roomCounter int64
	Database    *sql.DB

	AdminMu    sync.Mutex
	AdminConns map[*websocket.Conn]struct{}

	JWTSecret []byte
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Hub) AdminWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			auth := r.Header.Get("Authorization")
			if len(auth) > 7 && auth[:7] == "Bearer " {
				tokenStr = auth[7:]
			}
		}
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return h.JWTSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("AdminWS upgrade failed:", err)
			return
		}

		h.AdminMu.Lock()
		if h.AdminConns == nil {
			h.AdminConns = make(map[*websocket.Conn]struct{})
		}
		h.AdminConns[conn] = struct{}{}
		h.AdminMu.Unlock()

		log.Println("AdminWS: admin client connected")

		go func() {
			defer func() {
				h.AdminMu.Lock()
				delete(h.AdminConns, conn)
				h.AdminMu.Unlock()
				conn.Close()
				log.Println("AdminWS: admin client disconnected")
			}()
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()
	}
}

func (h *Hub) CreateRoom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("CreateRoom: Received request")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			log.Println("CreateRoom read body error:", err)
			return
		}
		defer r.Body.Close()

		var newRoom models.Room
		if err := json.Unmarshal(body, &newRoom); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			log.Println("CreateRoom JSON unmarshal error:", err)
			return
		}
		log.Printf("CreateRoom: Room name = %s\n", newRoom.RoomName)

		h.Mu.Lock()
		h.roomCounter++
		roomID := h.roomCounter
		room := &models.Room{
			RoomName: newRoom.RoomName,
			Users:    []int64{},
		}
		h.Cache.Rooms[roomID] = room
		h.Mu.Unlock()

		fetchGames := func(userID, profileID int64) ([]models.Game, error) {
			return profiles.GetGamesByUserAndProfile(h.Database, userID, profileID)
		}

		tasksFilepath, err := getconfig.GetTasksFilePath()
		if err != nil {
			log.Println("Failed to load tasks file path:", err)
			tasksFilepath = ""
		}

		fetchTask := func() (string, error) {
			if tasksFilepath == "" {
				return "", errors.New("tasks file path not configured")
			}
			type tasksPayload struct {
				Tasks []string `json:"tasks"`
			}
			b, err := os.ReadFile(tasksFilepath)
			if err != nil {
				return "", fmt.Errorf("read tasks file: %w", err)
			}
			var tp tasksPayload
			if err := json.Unmarshal(b, &tp); err != nil {
				return "", fmt.Errorf("parse tasks file: %w", err)
			}
			if len(tp.Tasks) == 0 {
				return "", errors.New("no tasks in tasks file")
			}
			rand.Seed(time.Now().UnixNano())
			return tp.Tasks[rand.Intn(len(tp.Tasks))], nil
		}

		roomServer := rooms.NewRoomServer(room, roomID, h.Cache, fetchGames, fetchTask)

		h.Mu.Lock()
		h.RoomServers[roomID] = roomServer
		h.Mu.Unlock()

		roomServer.Start()

		log.Printf("CreateRoom: Room %d created\n", roomID)

		update := models.RoomUpdate{
			Action: "create",
			RoomID: roomID,
			Name:   newRoom.RoomName,
		}
		select {
		case h.Cache.Broadcast <- update:
		default:
			log.Printf("Broadcast channel full or no listener; skipped create update for room %d", roomID)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"roomID": %d, "roomName": "%s"}`, roomID, newRoom.RoomName)))
	}
}

func (h *Hub) DeleteRoom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("DeleteRoom: Received request")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			log.Println("DeleteRoom read body error:", err)
			return
		}
		defer r.Body.Close()

		var req struct {
			RoomID int64 `json:"room_id"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			log.Println("DeleteRoom JSON unmarshal error:", err)
			return
		}

		h.Mu.Lock()
		srv, exists := h.RoomServers[req.RoomID]
		if exists {
			delete(h.RoomServers, req.RoomID)
			delete(h.Cache.Rooms, req.RoomID)
		}
		h.Mu.Unlock()

		if !exists {
			http.Error(w, "Room not found", http.StatusNotFound)
			log.Printf("DeleteRoom: Room %d not found\n", req.RoomID)
			return
		}

		srv.Stop()
		log.Printf("DeleteRoom: Room %d stopped and deleted\n", req.RoomID)

		update := models.RoomUpdate{
			Action: "delete",
			RoomID: req.RoomID,
		}
		select {
		case h.Cache.Broadcast <- update:
		default:
			log.Printf("Broadcast channel full or no listener; skipped delete update for room %d", req.RoomID)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Room deleted"}`))
	}
}

func (h *Hub) GetRooms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Mu.Lock()
		defer h.Mu.Unlock()

		type roomInfo struct {
			RoomID   int64   `json:"room_id"`
			RoomName string  `json:"room_name"`
			Users    []int64 `json:"users"`
		}

		rooms := make([]roomInfo, 0, len(h.Cache.Rooms))
		for id, room := range h.Cache.Rooms {
			rooms = append(rooms, roomInfo{
				RoomID:   id,
				RoomName: room.RoomName,
				Users:    room.GetUsers(),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(rooms); err != nil {
			http.Error(w, "Failed to encode rooms", http.StatusInternalServerError)
		}
	}
}

func StartHub() error {
	log.Println("Starting Hub...")

	portInt, err := getconfig.GetHubPort()
	if err != nil {
		return err
	}
	jwtPath, err := getconfig.GetJWTDatabasePath()
	if err != nil {
		return err
	}

	Database, err := db.OpenDataBase(jwtPath)
	if err != nil {
		return err
	}

	secret, err := db.GetCurrentJWTKey(Database)
	if err != nil {
		return err
	}

	userDBPath, err := getconfig.GetProfilsDatabasePath()
	if err != nil {
		return err
	}

	DatabaseNew, err := db.OpenDataBase(userDBPath)
	if err != nil {
		return err
	}

	hub := &Hub{
		Cache: &models.HubCache{
			Rooms:        make(map[int64]*models.Room),
			Broadcast:    make(chan models.RoomUpdate, 64),
			UserProfiles: make(map[int64]models.ProfileData),
		},
		RoomServers: make(map[int64]*rooms.RoomServer),
		Database:    DatabaseNew,
		AdminConns:  make(map[*websocket.Conn]struct{}),
		JWTSecret:   []byte(secret),
	}

	go func() {
		for upd := range hub.Cache.Broadcast {
			b, err := json.Marshal(upd)
			if err != nil {
				log.Println("failed to marshal broadcast:", err)
				continue
			}
			hub.AdminMu.Lock()
			for c := range hub.AdminConns {
				if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Printf("admin ws write error: %v — removing conn\n", err)
					c.Close()
					delete(hub.AdminConns, c)
				}
			}
			hub.AdminMu.Unlock()
			log.Printf("Hub broadcast: %+v\n", upd)
		}
	}()

	http.Handle("/getRooms", enableCORS(hub.GetRooms()))
	http.Handle("/createRoom", enableCORS(hub.CreateRoom()))
	http.Handle("/deleteRoom", enableCORS(hub.DeleteRoom()))
	http.Handle("/admin/ws", enableCORS(hub.AdminWS()))

	sessionServer := session.NewSessionServer(hub.Cache, hub.RoomServers, hub.Database, []byte(secret))
	http.Handle("/ws", enableCORS(http.HandlerFunc(sessionServer.HandleWS)))

	log.Printf("Hub listening on port %d\n", portInt)
	return http.ListenAndServe(":"+strconv.Itoa(portInt), nil)
}
