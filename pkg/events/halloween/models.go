package halloween

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Main Cache where all Tournament data is stored
type Cache struct {
	Tournaments []Tournament
	Mu          sync.Mutex
}

// Tournament Server cache
type Tournament struct {
	Name    string
	Id      int64
	Teams   []Team
	Rounds  []Round
	Players []Client
}

// Client struct with the connection
type Client struct {
	Conn websocket.Conn
	Id   int64
}

// necessary structures
type Team struct {
	Players       []int64
	Name          string
	Id            int64
	PumpkinHealth int
}

type Round struct {
	Game Game
	Id   int64
}

// Power is how much the win will deal damage to another pumpkin
type Game struct {
	Name  string
	Power int
}

// Messages Types
type JoinMessage struct {
	UserID  int64
	Message string
}

type LeaveMessage struct {
	UserId  int64
	Message string
}

type BroadcastMassage struct {
	Message string
	Type    string
}

// Halloween Server Cache
type HalloweenServer struct {
	Join      chan JoinMessage
	Leave     chan LeaveMessage
	Broadcast chan BroadcastMassage
}
