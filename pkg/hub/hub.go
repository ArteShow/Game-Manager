package hub

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ArteShow/Game-Manager/models"
	getconfig "github.com/ArteShow/Game-Manager/pkg/getConfig"
)

func CreateRoom(hubCache *models.HubCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read the body", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		defer r.Body.Close()

		var NewRoom models.Room
		err = json.Unmarshal(body, &NewRoom)
		if err != nil {
			http.Error(w, "Failed to unmarshal the body", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		roomID := int64(1)
		for {
			if _, exists := hubCache.Rooms[roomID]; !exists {
				break
			}
			roomID++
		}

		hubCache.Start <- models.StartRoomRequest{
			RoomID:   roomID,
			RoomName: NewRoom.RoomName,
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"roomID": %d, "roomName": "%s"}`, roomID, NewRoom.RoomName)))
	}
}

func DeleteRoom(hubCache *models.HubCache) http.HandlerFunc {
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

		if _, exists := hubCache.Rooms[req.RoomID]; exists {
			delete(hubCache.Rooms, req.RoomID)
			log.Printf("Room %d deleted\n", req.RoomID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "Room deleted"}`))
		} else {
			http.Error(w, "Room not found", http.StatusNotFound)
		}
	}
}

func StartHub() error {
	log.Println("Starting the Hub")

	hubCache := &models.HubCache{
		Rooms: make(map[int64]*models.Room),
		Join:  make(chan models.JoinRequest),
		Leave: make(chan models.LeaveRequest),
		Start: make(chan models.StartRoomRequest),
	}

	portInt, err := getconfig.GetHubPort()
	if err != nil {
		return err
	}
	portStr := ":" + strconv.Itoa(portInt)

	go func() {
		for {
			select {
			case joinReq := <-hubCache.Join:
				room, ok := hubCache.Rooms[joinReq.RoomID]
				if !ok {
					log.Printf("Room %d does not exist\n", joinReq.RoomID)
					continue
				}
				room.AddUser(joinReq.UserID)
				log.Printf("User %d joined room %d\n", joinReq.UserID, joinReq.RoomID)

			case leaveReq := <-hubCache.Leave:
				room, ok := hubCache.Rooms[leaveReq.RoomID]
				if !ok {
					log.Printf("Room %d does not exist\n", leaveReq.RoomID)
					continue
				}
				room.RemoveUser(leaveReq.UserID)
				log.Printf("User %d left room %d\n", leaveReq.UserID, leaveReq.RoomID)

			case startReq := <-hubCache.Start:
				if _, exists := hubCache.Rooms[startReq.RoomID]; !exists {
					hubCache.Rooms[startReq.RoomID] = &models.Room{
						RoomName: startReq.RoomName,
						Users:    []int64{},
						Started:  true,
					}
					log.Printf("Room %d started: %s\n", startReq.RoomID, startReq.RoomName)
				}
			}
		}
	}()

	http.HandleFunc("/createRoom", CreateRoom(hubCache))
	http.HandleFunc("/deleteRoom", DeleteRoom(hubCache))

	return http.ListenAndServe(portStr, nil)
}
