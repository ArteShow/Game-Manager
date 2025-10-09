package tournament

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Main Cache where all Tournament data is storted
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

// nesecery structures
type Team struct {
	Players []int64
	Name    string
	Id      int64
}

type Round struct {
	Games string
	Id    int64
}
