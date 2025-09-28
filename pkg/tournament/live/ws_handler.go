package live

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/ArteShow/Game-Manager/models"
)

type WsServer struct {
	Server models.LiveServer
	Mu     sync.Mutex
}

func NewWsServer(NewLiveServer *models.LiveServer) *WsServer {
	return &WsServer{
		Server: *NewLiveServer,
		Mu:     sync.Mutex{},
	}
}

func (ws *WsServer) StartWs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(string)
		if !ok {
			http.Error(w, "userID not found", http.StatusUnauthorized)
			return
		}
		intUserId, err := strconv.Atoi(userID)
		if err != nil {
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		NewUser := models.Client{
			Conn:          conn,
			UserId:        int64(intUserId),
			ClientMessage: make(chan models.ClientMessage),
		}

		ws.Mu.Lock()
		ws.Server.Clients = append(ws.Server.Clients, NewUser)
		ws.Mu.Unlock()

		go NewUser.WritePump(ws.Server)
		go NewUser.ReadPump(&ws.Server)
	}
}
