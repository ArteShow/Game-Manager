package halloween

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // adjust for your needs
}

func HalloweenWebsocketServer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	//Save connection
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusBadRequest)
		return
	}

	//Start Client Functions
	go ReadPump(conn)
	go WritePump(conn, userID)
}

func ReadPump(conn *websocket.Conn) {
	for {
		//Listen for Message
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		type MessageType struct {
			Type string `json:"type"`
		}
		var MType MessageType

		err = json.Unmarshal(message, &MType)
		if err != nil {
			log.Println("Failed to read the message")
			return
		}
		//Hear Messages from the user
		//If type is JOIN
		if MType.Type == "JOIN" {
			var JoinMessage JoinMessage
			err := json.Unmarshal(message, &JoinMessage)
			if err != nil {
				log.Println("Failed to read the message 2")
			}

			//Look for the right hw server and send the data in the right chanel
			for _, hw := range cache.HalloweenGame {
				if hw.Id == JoinMessage.HalloweenGameId {
					for _, hws := range cache.HalloweenServers {
						if hws.Id == hw.Id {
							hws.Join <- JoinMessage
						}
					}
				}
			}
		} else if MType.Type == "LEAVE" {
			//Get the message
			var LeaveMessage LeaveMessage
			err := json.Unmarshal(message, &LeaveMessage)
			if err != nil {
				log.Println("Failed to read the message")
			}
			LeaveMessage.Type = "LEAVE"
			//Look for the right channel to send the info
			for _, hw := range cache.HalloweenGame {
				for _, cl := range hw.Players {
					if cl.Id == LeaveMessage.UserId {
						for _, hws := range cache.HalloweenServers {
							if hws.Id == hw.Id {
								hws.Leave <- LeaveMessage
							}
						}
					}
				}
			}
		}
	}
}

func WritePump(conn *websocket.Conn, userID int64) {
	//Look for the game
	for {
		for _, hw := range cache.HalloweenGame {
			for _, cl := range hw.Players {
				if cl.Id == userID {
					//Look for the Halloween Server
					for _, hws := range cache.HalloweenServers {
						if hws.Id == hw.Id {
							//Hear check for server events
							select {
							case msg := <-hws.Broadcast:
								err := cl.Conn.WriteJSON(msg)
								if err != nil {
									log.Println("Failed to send the broadcast message")
									return
								}
							}
						}
					}
				}
			}
		}
	}
}
