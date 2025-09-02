package hub

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/ArteShow/Game-Manager/models"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
	"github.com/ArteShow/Game-Manager/pkg/rooms"
)

type Hub struct {
	Cache       *models.HubCache
	RoomServers map[int64]*rooms.RoomServer
	Mu          *sync.Mutex
}

func (h *Hub) CreateRoom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		defer r.Body.Close()

		var newRoom models.Room
		if err := json.Unmarshal(body, &newRoom); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			log.Println(err)
			return
		}

		h.Mu.Lock()
		roomID := int64(len(h.Cache.Rooms) + 1)

		room := &models.Room{
			RoomName: newRoom.RoomName,
			Users:    []int64{},
		}
		h.Cache.Rooms[roomID] = room

		roomServer := rooms.NewRoomServer(room, roomID, h.Cache)
		h.RoomServers[roomID] = roomServer
		roomServer.Start()
		h.Mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"roomID": %d, "roomName": "%s"}`, roomID, newRoom.RoomName)))
	}
}

func (h *Hub) DeleteRoom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		defer r.Body.Close()

		var req struct {
			RoomID int64 `json:"room_id"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			log.Println(err)
			return
		}

		h.Mu.Lock()
		if _, exists := h.Cache.Rooms[req.RoomID]; exists {
			if srv, ok := h.RoomServers[req.RoomID]; ok {
				srv.Stop()
				delete(h.RoomServers, req.RoomID)
			}

			delete(h.Cache.Rooms, req.RoomID)
			log.Printf("Room %d deleted\n", req.RoomID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "Room deleted"}`))
		} else {
			http.Error(w, "Room not found", http.StatusNotFound)
		}
		h.Mu.Unlock()
	}
}

func StartHub() error {
	log.Println("Starting the Hub")

	hub := &Hub{
		Cache: &models.HubCache{
			Rooms: make(map[int64]*models.Room),
			Join:  make(chan models.JoinRequest),
			Leave: make(chan models.LeaveRequest),
			Start: make(chan models.StartRoomRequest),
		},
		RoomServers: make(map[int64]*rooms.RoomServer),
		Mu:          &sync.Mutex{},
	}

	portInt, err := getconfig.GetHubPort()
	if err != nil {
		return err
	}
	portStr := ":" + strconv.Itoa(portInt)

	go func() {
		for {
			select {
			case joinReq := <-hub.Cache.Join:
				room, ok := hub.Cache.Rooms[joinReq.RoomID]
				if !ok {
					log.Printf("Room %d does not exist\n", joinReq.RoomID)
					continue
				}
				room.AddUser(joinReq.UserID)
				log.Printf("User %d joined room %d\n", joinReq.UserID, joinReq.RoomID)

			case leaveReq := <-hub.Cache.Leave:
				room, ok := hub.Cache.Rooms[leaveReq.RoomID]
				if !ok {
					log.Printf("Room %d does not exist\n", leaveReq.RoomID)
					continue
				}
				room.RemoveUser(leaveReq.UserID)
				log.Printf("User %d left room %d\n", leaveReq.UserID, leaveReq.RoomID)

			case startReq := <-hub.Cache.Start:
				if _, exists := hub.Cache.Rooms[startReq.RoomID]; !exists {
					hub.Cache.Rooms[startReq.RoomID] = &models.Room{
						RoomName: startReq.RoomName,
						Users:    []int64{},
					}
					log.Printf("Room %d started: %s\n", startReq.RoomID, startReq.RoomName)
				}
			}
		}
	}()

	http.HandleFunc("/createRoom", hub.CreateRoom())
	http.HandleFunc("/deleteRoom", hub.DeleteRoom())

	return http.ListenAndServe(portStr, nil)
}
