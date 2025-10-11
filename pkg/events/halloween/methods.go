package halloween

// Get max id of hw
func (c *Cache) GetMaxId() int64 {
	var max int64
	for _, t := range c.HalloweenGame {
		if max < t.Id {
			max = t.Id
		}
	}

	return max
}

// Add hw to the cache
func (c *Cache) AddHalloweenGame(client Client, Name string) int64 {
	c.Mu.Lock()

	MaxId := c.GetMaxId()
	c.HalloweenGame = append(c.HalloweenGame, HalloweenGame{
		Name:    Name,
		Id:      c.GetMaxId(),
		Teams:   []Team{},
		Rounds:  []Round{},
		Players: []Client{},
		Admin:   client.Id,
	})
	c.Mu.Unlock()

	go StartHalloweenGameServer(MaxId)

	return MaxId
}

// delete hw from cache
func (c *Cache) DeleteHalloweenGame(id int64) {
	c.Mu.Lock()
	for i, t := range c.HalloweenGame {
		if t.Id == id {
			c.HalloweenGame = append(c.HalloweenGame[:i], c.HalloweenGame[i+1:]...)
		}
	}
	c.Mu.Unlock()
}

// get all hw from cache
func (c *Cache) GetHalloweenGames() []HalloweenGame {
	var Halloweengames []HalloweenGame
	for _, h := range c.HalloweenGame {
		Halloweengames = append(Halloweengames, h)
	}

	return Halloweengames
}
