package halloween

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

func StartHalloweenGameServer(hwGameId int64) {
	//Initialize HWServerCache
	HWServerCache := HalloweenServer{
		Join:      make(chan JoinMessage),
		Leave:     make(chan LeaveMessage),
		Broadcast: make(chan BroadcastMassage),
		Id:        hwGameId,
	}

	cache.HalloweenServers = append(cache.HalloweenServers, HWServerCache)

	//Start the infinite loop
	for _, hw := range cache.HalloweenGame {
		if hw.Id == hwGameId {
			for {
				//Check for messages in the channels
				select {
				case msg := <-HWServerCache.Join:
					if msg.Message == "JOIN" {
						//Check if user is already in a game
						var counter int
						for _, HW := range cache.HalloweenGame {
							for _, cl := range HW.Players {
								if cl.Id == msg.UserID {
									counter++
								}
							}
						}

						if counter < 1 {
							cache.Mu.Lock()
							hw.Players = append(hw.Players, Client{Conn: *msg.Conn, Id: msg.UserID})
							cache.Mu.Unlock()
						} else {
							msg.Conn.WriteJSON(BroadcastMassage{Message: "You are already in a game", Type: "ERROR"})
						}

						//Broadcasting all other users
						stringID := strconv.Itoa(int(msg.UserID))
						Broadcastmessage := BroadcastMassage{
							Message: stringID,
							Type:    "JOIN",
						}

						HWServerCache.Broadcast <- Broadcastmessage
					}
				//If stop then stop lol
				case msg := <-HWServerCache.Stop:
					if msg.Type == "STOP" {
						for _, hw := range cache.HalloweenGame {
							if HWServerCache.Id == hw.Id {
								if hw.Admin == msg.AdminID {
									break
								} else {
									//write error as string message
									err2 := msg.Conn.WriteMessage(websocket.TextMessage, []byte("You are not an admin"))
									if err2 != nil {
										log.Println(err2)
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
}
