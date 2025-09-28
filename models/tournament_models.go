package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type LiveServer struct {
	Tournaments       []Tournament
	Mu                sync.Mutex
	Broadcast         chan BroadcastMessage
	Join              chan JoinChan
	Clients           []Client
	TournamentServers []TournamentServer
}

type Client struct {
	Conn          *websocket.Conn
	UserId        int64
	ClientMessage chan ClientMessage
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
	Message string `json:"message"`
	UserId  int64  `json:"id"`
}

type TournamentMessage struct {
	UserId  int64
	Message string
}

type ClientMessage struct {
	Message string `json:"message"`
}

type JoinChan struct {
	UserID  int64
	Message string
}

type LeaveChan struct {
	UserID int64
}

type TournamentServer struct {
	Join       chan JoinChan
	Leave      chan LeaveChan
	BrodCast   chan BroadcastMessage
	Stop       chan string
	Tournament Tournament
	ID         int64
	Mu         sync.Mutex
}
