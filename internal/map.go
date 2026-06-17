package internal

import (
	"sync"
	"time"

	"github.com/streame-gg/go-discord-wrapper/types/discord"
)

type MessageCache struct {
	mu   sync.Mutex
	data map[discord.Snowflake]map[discord.Snowflake][]discord.Snowflake
}

func NewMessageCache() *MessageCache {
	return &MessageCache{
		data: make(map[discord.Snowflake]map[discord.Snowflake][]discord.Snowflake),
	}
}

// Add fügt eine MessageID unter userID -> channelID hinzu und entfernt sie nach 5 Minuten automatisch wieder
func (c *MessageCache) Add(userID, channelID, msgID discord.Snowflake) {
	c.mu.Lock()
	if _, ok := c.data[userID]; !ok {
		c.data[userID] = make(map[discord.Snowflake][]discord.Snowflake)
	}
	c.data[userID][channelID] = append(c.data[userID][channelID], msgID)
	c.mu.Unlock()

	time.AfterFunc(5*time.Minute, func() {
		c.remove(userID, channelID, msgID)
	})
}

// remove entfernt eine einzelne MessageID und räumt leere Maps/Slices auf
func (c *MessageCache) remove(userID, channelID, msgID discord.Snowflake) {
	c.mu.Lock()
	defer c.mu.Unlock()

	channels, ok := c.data[userID]
	if !ok {
		return
	}

	msgs, ok := channels[channelID]
	if !ok {
		return
	}

	for i, id := range msgs {
		if id == msgID {
			channels[channelID] = append(msgs[:i], msgs[i+1:]...)
			break
		}
	}

	if len(channels[channelID]) == 0 {
		delete(channels, channelID)
	}
	if len(channels) == 0 {
		delete(c.data, userID)
	}
}

// Get liest alle gecachten MessageIDs für eine userID + channelID
func (c *MessageCache) Get(userID discord.Snowflake) map[discord.Snowflake][]discord.Snowflake {
	c.mu.Lock()
	defer c.mu.Unlock()

	channels, ok := c.data[userID]
	if !ok {
		return nil
	}
	return channels
}
