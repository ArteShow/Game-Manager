package halloween

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Main Cache where all Tournament data is stored
type Cache struct {
	HalloweenGame    []HalloweenGame
	Mu               sync.Mutex
	HalloweenServers []HalloweenServer
}

// Tournament Server cache
type HalloweenGame struct {
	Name    string   `json:"name"`
	Id      int64    `json:"id"`
	Teams   []Team   `json:"teams"`
	Rounds  []Round  `json:"rounds"`
	Players []Client `json:"players"`
	Admin   int64    `json:"admin"`
}

// Client struct with the connection
type Client struct {
	Conn websocket.Conn
	Id   int64 `json:"id"`
}

// necessary structures
type Team struct {
	Players       []int64 `json:"players"`
	Name          string  `json:"name"`
	Id            int64   `json:"id"`
	PumpkinHealth int     `json:"health"`
}

type Round struct {
	Game Game  `json:"game"`
	Id   int64 `json:"id"`
}

// Power is how much the win will deal damage to another pumpkin
type Game struct {
	Name  string `json:"name"`
	Power int    `json:"power"`
}

// Messages Types
type JoinMessage struct {
	UserID          int64  `json:"userId"`
	Message         string `json:"message"`
	Conn            *websocket.Conn
	HalloweenGameId int64 `json:"hw_id"`
}

type LeaveMessage struct {
	UserId int64  `json:"id"`
	Type   string `json:"type"`
}

type BroadcastMassage struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type StopMessage struct {
	AdminID int64
	Conn    *websocket.Conn
	Type    string
}

// Halloween Server Cache
type HalloweenServer struct {
	Join      chan JoinMessage
	Leave     chan LeaveMessage
	Broadcast chan BroadcastMassage
	Stop      chan StopMessage
	Id        int64
}
