package halloween

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // adjust for your needs
}

func HalloweenWebsocketServer(w http.ResponseWriter, r *http.Request) {
	//Get UserId from context
	userId := r.Context().Value("userID").(int64)

	//Get HWGame Id from the request
	type ClientRequest struct {
		ID int64 `json:"id"`
	}
	var req ClientRequest

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to read the body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Failed to unmarshal the body", http.StatusInternalServerError)
		return
	}

	//Save connection
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusBadRequest)
		return
	}

	for _, hw := range cache.HalloweenGame {
		for _, cl := range hw.Players {
			if cl.Id == req.ID {
				cache.Mu.Lock()
				cl.Conn = *conn
				cache.Mu.Unlock()
			}
		}
	}

	//Start Client Functions
	go ReadPump(userId, req.ID)
	go WritePump(userId, req.ID)
}

func ReadPump(userID, HalloweenGameId int64) {
	//Hear Wait for user Messages
}

func WritePump(userID, HalloweenGameId int64) {
	//Hear wait for server messages
}
