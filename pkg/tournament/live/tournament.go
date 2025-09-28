package live

import (
	"log"

	"github.com/ArteShow/Game-Manager/models"
)

func StartTournament(Ts models.TournamentServer) {
	go func() {
		for {
			select {
			case msg := <-Ts.Join:
				if msg.Message == "JOIN" {
					Ts.Tournament.Players = append(Ts.Tournament.Players, msg.UserID)
					log.Println("New Player joined!")
				}
			}
		}
	}()
}
