package rooms

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/ArteShow/Game-Manager/models"
	"github.com/gorilla/websocket"
)

type RoomServer struct {
	Room       *models.Room
	RoomID     int64
	HubCache   *models.HubCache
	Join       chan int64
	Leave      chan int64
	Messages   chan models.RoomMessage
	stop       chan bool
	FetchGames func(userID, profileID int64) ([]models.Game, error)
	FetchTask  func() (string, error)

	mu    sync.Mutex
	conns map[int64]*websocket.Conn
}

func NewRoomServer(
	room *models.Room,
	roomID int64,
	hub *models.HubCache,
	fetchGames func(userID, profileID int64) ([]models.Game, error),
	fetchTask func() (string, error),
) *RoomServer {
	return &RoomServer{
		Room:       room,
		RoomID:     roomID,
		HubCache:   hub,
		Join:       make(chan int64),
		Leave:      make(chan int64),
		Messages:   make(chan models.RoomMessage),
		stop:       make(chan bool),
		FetchGames: fetchGames,
		FetchTask:  fetchTask,
		conns:      make(map[int64]*websocket.Conn),
	}
}

func (r *RoomServer) AddConnection(userID int64, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[userID] = conn
	log.Printf("Room %d: User %d connected\n", r.RoomID, userID)
}

func (r *RoomServer) RemoveConnection(userID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, userID)
	log.Printf("Room %d: User %d connection removed\n", r.RoomID, userID)
}

func (r *RoomServer) broadcastToRoom(payload interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	data, err := json.Marshal(payload)
	if err != nil {
		log.Println("broadcast marshal error:", err)
		return
	}
	for userID, conn := range r.conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Room %d: failed to send to user %d: %v\n", r.RoomID, userID, err)
		}
	}
}

func (r *RoomServer) chooseRandomGame() (models.Game, error) {
	userIDs := r.Room.GetUsers()
	gameSet := make(map[string]models.Game)
	for _, userID := range userIDs {
		profile, ok := r.HubCache.UserProfiles[userID]
		if !ok {
			log.Printf("Room %d: No profile cached for user %d\n", r.RoomID, userID)
			continue
		}
		games, err := r.FetchGames(userID, profile.ProfileID)
		if err != nil {
			log.Printf("Room %d: fetch games for user %d failed: %v", r.RoomID, userID, err)
			continue
		}
		for _, g := range games {
			key := fmt.Sprintf("%d:%s", g.ProfileID, g.Name)
			gameSet[key] = g
		}
	}
	allGames := make([]models.Game, 0, len(gameSet))
	for _, g := range gameSet {
		allGames = append(allGames, g)
	}
	if len(allGames) == 0 {
		return models.Game{}, fmt.Errorf("no games available in room %d", r.RoomID)
	}
	rand.Seed(time.Now().UnixNano())
	chosen := allGames[rand.Intn(len(allGames))]
	log.Printf("Room %d: Random game chosen: %v\n", r.RoomID, chosen)
	return chosen, nil
}

func (r *RoomServer) chooseRandomTask() (string, error) {
	task, err := r.FetchTask()
	if err != nil {
		log.Printf("Room %d: Failed to choose task: %v\n", r.RoomID, err)
	} else {
		log.Printf("Room %d: Random task chosen: %s\n", r.RoomID, task)
	}
	return task, err
}

func (r *RoomServer) Start() {
	go func() {
		log.Printf("Room %d started: %s\n", r.RoomID, r.Room.RoomName)
		for {
			select {
			case userID := <-r.Join:
				r.Room.AddUser(userID)
				log.Printf("Room %d: User %d joined, total users: %v\n", r.RoomID, userID, r.Room.GetUsers())
				r.broadcastToRoom(map[string]interface{}{
					"type":  "user_joined",
					"user":  userID,
					"users": r.Room.GetUsers(),
				})

			case userID := <-r.Leave:
				r.Room.RemoveUser(userID)
				r.RemoveConnection(userID)
				log.Printf("Room %d: User %d left, total users: %v\n", r.RoomID, userID, r.Room.GetUsers())
				r.broadcastToRoom(map[string]interface{}{
					"type":  "user_left",
					"user":  userID,
					"users": r.Room.GetUsers(),
				})

			case msg := <-r.Messages:
				log.Printf("Room %d: Message from user %d: %s\n", r.RoomID, msg.UserID, msg.Message)

				if msg.Message == "START" {
					game, err := r.chooseRandomGame()
					if err != nil {
						r.broadcastToRoom(map[string]interface{}{"type": "error", "error": err.Error()})
						continue
					}
					r.broadcastToRoom(map[string]interface{}{"type": "game_chosen", "game": game})
				}

				if msg.Message == "TASK" {
					task, err := r.chooseRandomTask()
					if err != nil {
						r.broadcastToRoom(map[string]interface{}{"type": "error", "error": err.Error()})
						continue
					}
					r.broadcastToRoom(map[string]interface{}{"type": "task_chosen", "task": task})
				}

			case <-r.stop:
				r.broadcastToRoom(map[string]interface{}{
					"type":  "room_closed",
					"room":  r.RoomID,
					"users": r.Room.GetUsers(),
				})
				log.Printf("Room %d stopped\n", r.RoomID)
				return
			}
		}
	}()
}

func (r *RoomServer) Stop() {
	log.Printf("Room %d: Stop signal received\n", r.RoomID)
	r.stop <- true
}
