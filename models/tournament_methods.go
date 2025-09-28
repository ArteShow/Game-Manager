package models

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

func (c *Client) ReadPump(server *LiveServer) {
	defer func() {
		server.Mu.Lock()
		for i, cl := range server.Clients {
			if cl == *c {
				server.Clients = append(server.Clients[:i], server.Clients[i+1:]...)
				break
			}
		}
		server.Mu.Unlock()
		c.Conn.Close()
	}()
	for {
		_, msg, _ := c.Conn.ReadMessage()
		var MSG ClientMessage
		err := json.Unmarshal(msg, &MSG)
		if err != nil {
			return
		}

		if MSG.Message == "JOIN" {
			server.Join <- JoinChan{UserID: c.UserId, Message: "JOIN"}
		}
	}
}

func (c *Client) WritePump(lv LiveServer) {
	defer c.Conn.Close()

	for {
		select {
		case msg := <-lv.Broadcast:
			if msg.Message == "JOIN" {
				bytes, err := json.Marshal(msg)
				if err != nil {
					return
				}

				c.Conn.WriteMessage(websocket.TextMessage, bytes)
			}
		}
	}
}

func (lv *LiveServer) GetMaxTournamentID() int64 {
	lv.Mu.Lock()
	defer lv.Mu.Unlock()
	var maxID int64 = 0
	for _, tournament := range lv.Tournaments {
		if tournament.ID > maxID {
			maxID = tournament.ID
		}
	}
	return maxID
}

func (lv *LiveServer) AddTournament() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(string)
		if !ok {
			http.Error(w, "userID not found", http.StatusUnauthorized)
			return
		}
		intUserId, err := strconv.Atoi(userID)
		if err != nil {
			return
		}
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read the body", http.StatusInternalServerError)
			return
		}

		var Rounds Tournament
		err = json.Unmarshal(body, &Rounds)
		if err != nil {
			http.Error(w, "Failed to encode the body", http.StatusInternalServerError)
			return
		}

		TournamentID := lv.GetMaxTournamentID() + 1

		NewTournament := Tournament{
			Teams:   []Team{},
			Rounds:  Rounds.Rounds,
			Players: []int64{int64(intUserId)},
			Admin:   int64(intUserId),
			Name:    Rounds.Name,
			ID:      TournamentID,
		}

		lv.Mu.Lock()
		defer lv.Mu.Unlock()
		lv.Tournaments = append(lv.Tournaments, NewTournament)

		response, err := json.Marshal(TournamentID)
		if err != nil {
			http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
			return
		}

		var NewServer TournamentServer
		NewServer.NewServer(NewTournament, TournamentID)
		go NewServer.StartTournament(*lv)

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
		w.WriteHeader(http.StatusAccepted)
	}
}

func (lv *LiveServer) DeleteTournament() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(string)
		if !ok {
			http.Error(w, "userID not found", http.StatusUnauthorized)
			return
		}
		intUserId, err := strconv.Atoi(userID)
		if err != nil {
			return
		}
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read the body", http.StatusInternalServerError)
			return
		}

		var req Tournament

		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Failed to encode the body", http.StatusInternalServerError)
			return
		}

		lv.Mu.Lock()
		defer lv.Mu.Unlock()
		for i, t := range lv.Tournaments {
			if t.ID == req.ID {
				if t.Admin != int64(intUserId) {
					http.Error(w, "Only the admin can delete the tournament", http.StatusUnauthorized)
					return
				}
				lv.Tournaments = append(lv.Tournaments[:i], lv.Tournaments[i+1:]...)
				w.WriteHeader(http.StatusAccepted)
				return
			}
		}

		for _, ts := range lv.TournamentServers {
			if ts.ID == req.ID {
				if ts.Tournament.Admin != int64(intUserId) {
					return
				} else {
					ts.Stop <- "STOP"
				}
			}
		}

		http.Error(w, "Tournament not found", http.StatusNotFound)
	}
}

func (lv *LiveServer) GetTournamets(w http.ResponseWriter, r *http.Request) {
	Bytes, err := json.Marshal(lv.Tournaments)
	if err != nil {
		http.Error(w, "Failed to get Tournaments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(Bytes)
	w.WriteHeader(http.StatusAccepted)
}

func (ts *TournamentServer) NewServer(Tr Tournament, id int64) TournamentServer {
	return TournamentServer{
		Join:       make(chan JoinChan),
		Leave:      make(chan LeaveChan),
		BrodCast:   make(chan BroadcastMessage),
		Stop:       make(chan string),
		Tournament: Tr,
		ID:         id,
	}
}

func (Ts *TournamentServer) StartTournament(lv LiveServer) {
	go func() {
		for {
			select {
			case msg := <-Ts.Join:
				if msg.Message == "JOIN" {
					Ts.Mu.Lock()
					Ts.Tournament.Players = append(Ts.Tournament.Players, msg.UserID)
					Ts.Mu.Unlock()

					lv.Broadcast <- BroadcastMessage{Message: "JOIN", UserId: msg.UserID}
				}
			case msg := <-Ts.Stop:
				if msg == "STOP" {
					for i, client := range lv.Clients {
						for _, id := range Ts.Tournament.Players {
							if id == client.UserId {
								lv.Mu.Lock()
								client.Conn.Close()
								lv.Clients = append(lv.Clients[:i], lv.Clients[i+1:]...)
								lv.Mu.Unlock()
							}
						}
					}
					for i, ts := range lv.TournamentServers {
						if ts.ID == Ts.ID {
							lv.Mu.Lock()
							lv.TournamentServers = append(lv.TournamentServers[:i], lv.TournamentServers[i+1:]...)
							lv.Mu.Unlock()
						}
					}
				}
			}
		}
	}()
}
