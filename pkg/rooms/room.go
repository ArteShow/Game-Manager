package rooms

import (
	"log"

	"github.com/ArteShow/Game-Manager/models"
)

type RoomServer struct {
	Room     *models.Room
	RoomID   int64
	HubCache *models.HubCache
	Join     chan int64
	Leave    chan int64
	Messages chan string
	stop     chan bool
}

func NewRoomServer(room *models.Room, roomID int64, hub *models.HubCache) *RoomServer {
	return &RoomServer{
		Room:     room,
		RoomID:   roomID,
		HubCache: hub,
		Join:     make(chan int64),
		Leave:    make(chan int64),
		Messages: make(chan string),
		stop:     make(chan bool),
	}
}

func (r *RoomServer) Start() {
	go func() {
		log.Printf("Room %d started: %s\n", r.RoomID, r.Room.RoomName)
		for {
			select {
			case userID := <-r.Join:
				r.Room.AddUser(userID)
				log.Printf("User %d joined room %d\n", userID, r.RoomID)

			case userID := <-r.Leave:
				r.Room.RemoveUser(userID)
				log.Printf("User %d left room %d\n", userID, r.RoomID)

			case msg := <-r.Messages:
				log.Printf("Room %d message: %s\n", r.RoomID, msg)

			case <-r.stop:
				log.Printf("Room %d stopped\n", r.RoomID)
				return
			}
		}
	}()
}

func (r *RoomServer) Stop() {
	r.stop <- true
}
