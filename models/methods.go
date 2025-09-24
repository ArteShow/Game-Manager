package models

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

func (h *HubCache) GetAllProfiles() []ProfileData {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	profiles := make([]ProfileData, 0, len(h.UserProfiles))
	for _, profile := range h.UserProfiles {
		profiles = append(profiles, profile)
	}
	return profiles
}

func (r *Room) AddUser(userID int64) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Users = append(r.Users, userID)
}

func (r *Room) RemoveUser(userID int64) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	for i, id := range r.Users {
		if id == userID {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			break
		}
	}
}

func (r *Room) GetUsers() []int64 {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	usersCopy := make([]int64, len(r.Users))
	copy(usersCopy, r.Users)
	return usersCopy
}

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
}

func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
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

		NewTournament := Tournament{
			Teams:   []Team{},
			Rounds:  Rounds.Rounds,
			Players: []int64{int64(intUserId)},
			Admin:   int64(intUserId),
			Name:    Rounds.Name,
		}

		lv.Tournaments = append(lv.Tournaments, NewTournament)

		w.WriteHeader(http.StatusAccepted)
	}
}
