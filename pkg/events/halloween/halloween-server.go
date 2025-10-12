package halloween

import "strconv"

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
						hw.Players = append(hw.Players, Client{Conn: *msg.Conn, Id: msg.UserID})

						//Broadcasting all other users
						stringID := strconv.Itoa(int(msg.UserID))
						Broadcastmessage := BroadcastMassage{
							Message: stringID,
							Type:    "JOIN",
						}

						HWServerCache.Broadcast <- Broadcastmessage
					}
				}
			}
		}
	}
}
