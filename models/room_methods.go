package models

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
