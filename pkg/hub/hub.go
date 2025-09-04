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
)

type Hub struct {
	Cache       *models.HubCache
	RoomServers map[int64]*rooms.RoomServer
	Mu          sync.Mutex
	roomCounter int64
	Database    *sql.DB
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
		h.RoomServers[roomID] = roomServer
		roomServer.Start()

		log.Printf("CreateRoom: Room %d created\n", roomID)
		h.Cache.Broadcast <- models.RoomUpdate{
			Action: "create",
			RoomID: roomID,
			Name:   newRoom.RoomName,
		}
		h.Mu.Unlock()

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
		defer h.Mu.Unlock()
		if srv, exists := h.RoomServers[req.RoomID]; exists {
			srv.Stop()
			delete(h.RoomServers, req.RoomID)
			delete(h.Cache.Rooms, req.RoomID)
			log.Printf("DeleteRoom: Room %d stopped and deleted\n", req.RoomID)

			h.Cache.Broadcast <- models.RoomUpdate{
				Action: "delete",
				RoomID: req.RoomID,
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message": "Room deleted"}`))
		} else {
			http.Error(w, "Room not found", http.StatusNotFound)
			log.Printf("DeleteRoom: Room %d not found\n", req.RoomID)
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
			Broadcast:    make(chan models.RoomUpdate),
			UserProfiles: make(map[int64]models.ProfileData),
		},
		RoomServers: make(map[int64]*rooms.RoomServer),
		Database:    DatabaseNew,
	}

	http.HandleFunc("/createRoom", hub.CreateRoom())
	http.HandleFunc("/deleteRoom", hub.DeleteRoom())

	sessionServer := session.NewSessionServer(hub.Cache, hub.RoomServers, hub.Database, []byte(secret))
	http.HandleFunc("/ws", sessionServer.HandleWS)

	log.Printf("Hub listening on port %d\n", portInt)
	return http.ListenAndServe(":"+strconv.Itoa(portInt), nil)
}
