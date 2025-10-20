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
		Setup:     make(chan SetUp),
		Id:        hwGameId,
	}

	cache.HalloweenServers = append(cache.HalloweenServers, HWServerCache)

	//Start the infinite loop
	for _, HalloweenGame := range cache.HalloweenGame {
		if HalloweenGame.Id == hwGameId {
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
							HalloweenGame.Players = append(HalloweenGame.Players, Client{Conn: *msg.Conn, Id: msg.UserID})
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
						for i, hw := range cache.HalloweenGame {
							if HWServerCache.Id == hw.Id {
								if hw.Admin == msg.AdminID {
									for j, hws := range cache.HalloweenServers {
										if hws.Id == hw.Id {
											cache.Mu.Lock()
											cache.HalloweenServers = append(cache.HalloweenServers[:j], cache.HalloweenServers[j+1:]...)
											cache.Mu.Unlock()
										}
									}
									cache.Mu.Lock()
									cache.HalloweenGame = append(cache.HalloweenGame[:i], cache.HalloweenGame[i+1:]...)
									cache.Mu.Unlock()
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
				//Delete the person from all games
				case msg := <-HWServerCache.Leave:
					if msg.Type == "LEAVE" {
						for _, hw := range cache.HalloweenGame {
							for i, cl := range hw.Players {
								if cl.Id == msg.UserId {
									cache.Mu.Lock()
									hw.Players = append(hw.Players[:i], hw.Players[i+1:]...)
									cache.Mu.Unlock()

									//Broadcasting all other users
									BrMessage := BroadcastMassage{
										Message: "Player " + strconv.Itoa(int(msg.UserId)) + " has left the game",
										Type:    "LEAVE",
									}

									HWServerCache.Broadcast <- BrMessage
								}
							}
						}
					}
				case msg := <-HWServerCache.Setup:
					if msg.Type == "CREATE_TEAM" {
						//Look for the right hws
						for _, hw := range cache.HalloweenGame {
							for _, cl := range hw.Players {
								if cl.Id == msg.PlayerID {
									//Create New Team with pumpkin health of 100
									cache.Mu.Lock()
									hw.Teams = append(hw.Teams, Team{
										Players:       []int64{},
										Name:          msg.TeamName,
										Id:            hw.GetMaxId() + int64(1),
										PumpkinHealth: 100, //Need to add a config field for this
									})
									cache.Mu.Unlock()

									//Broadcasting everyone
									BrMessage := BroadcastMassage{
										Message: "New Team with Id: " + strconv.Itoa(int(hw.GetMaxId())+1) + "was created",
										Type:    "CREATE_TEAM",
									}

									HWServerCache.Broadcast <- BrMessage
								}
							}
						}
					}
				}
			}
		}
	}
}
