package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type LiveServer struct {
	Tournaments []Tournament
	Mu          sync.Mutex
	Broadcast chan BroadcastMessage
	Tournament chan TournamentMessage
	Clients []Client
}

type Client struct{
	Conn *websocket.Conn
	UserId int64
	Send chan []byte
}

type Tournament struct {
	Players []int64
	Name    string
	Rounds  []Round
	Teams []Team
}

type Team struct{
	Players []int64
	Points int
}

type Round struct {
	Games TournamentGame
}

type TournamentGame struct {
	Name string
}

type BroadcastMessage struct{
	Message string
}

type TournamentMessage struct{
	UserId int64
	Message string
}