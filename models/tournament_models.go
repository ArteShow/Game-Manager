package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type LiveServer struct {
	Tournaments []Tournament
	Mu          sync.Mutex
	Broadcast   chan BroadcastMessage
	Tournament  chan TournamentMessage
	Clients     []Client
}

type Client struct {
	Conn   *websocket.Conn
	UserId int64
	Send   chan []byte
}

type Tournament struct {
	Players []int64
	Name    string  `json:"name"`
	Rounds  []Round `json:"rounds"`
	Teams   []Team
	Admin   int64
	ID      int64 `json:"id"`
}

type Team struct {
	Players []int64
	Points  int
}

type Round struct {
	Games TournamentGame `json:"games"`
}

type TournamentGame struct {
	Name string `json:"game_name"`
}

type BroadcastMessage struct {
	Message string
}

type TournamentMessage struct {
	UserId  int64
	Message string
}
