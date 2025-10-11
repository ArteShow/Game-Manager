package halloween

func (c *Cache) GetMaxId() int64 {
	var max int64
	for _, t := range c.HalloweenGame {
		if max < t.Id {
			max = t.Id
		}
	}

	return max
}

func (c *Cache) AddHalloweenGame(client Client, Name string) {
	c.Mu.Lock()
	c.HalloweenGame = append(c.HalloweenGame, HalloweenGame{
		Name:    Name,
		Id:      c.GetMaxId(),
		Teams:   []Team{},
		Rounds:  []Round{},
		Players: []Client{client},
	})
	c.Mu.Unlock()
}

func (c *Cache) DeleteHalloweenGame(id int64) {
	c.Mu.Lock()
	for i, t := range c.HalloweenGame {
		if t.Id == id {
			c.HalloweenGame = append(c.HalloweenGame[:i], c.HalloweenGame[i+1:]...)
		}
	}
	c.Mu.Unlock()
}

func (c *Cache) GetHalloweenGames() []HalloweenGame {
	var Halloweengames []HalloweenGame
	for _, h := range c.HalloweenGame {
		Halloweengames = append(Halloweengames, h)
	}

	return Halloweengames
}
