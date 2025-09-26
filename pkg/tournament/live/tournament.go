package live

import (
	"log"

	"github.com/ArteShow/Game-Manager/models"
)

type TournamentServer struct {
	Join       chan models.JoinChan
	Leave      chan models.LeaveChan
	BroadCast  chan models.BroadcastMessage
	Tournament models.Tournament
}

func (t *TournamentServer) NewTouranemntServer() TournamentServer {
	return TournamentServer{
		Join:       make(chan models.JoinChan),
		Leave:      make(chan models.LeaveChan),
		BroadCast:  make(chan models.BroadcastMessage),
		Tournament: t.Tournament,
	}
}

func (t *TournamentServer) StartTournamentServer() {
	go func() {
		log.Println("Started room")
		for {
			select {
			case id := <-t.Join:
				if id.Message == "JOIN" {
					t.Tournament.Players = append(t.Tournament.Players, id.UserID)
					log.Println("New User joined")

					t.BroadCast <- models.BroadcastMessage{Message: "New User Joined"}
				}
			}
		}
	}()
}
