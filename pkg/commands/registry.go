package commands

import (
	"context"
	"fmt"
	"sync"

	"github.com/streame-gg/go-discord-wrapper/connection"
	"github.com/streame-gg/go-discord-wrapper/types/discord"
	"github.com/streame-gg/go-discord-wrapper/types/events"
)

type Command interface {
	Definition() *discord.ApplicationCommand
	Handle(c *connection.Client, ev *events.InteractionCreateEvent)
}

var (
	mu      sync.RWMutex
	sources []Command
	live    = map[string]Command{}
)

func Register(cmd Command) {
	sources = append(sources, cmd)
}

// Reload rebuilds the live command table from the registered sources. A
// duplicate-name error leaves the previous table intact, mirroring the events
// and components registries.
func Reload() error {
	m := make(map[string]Command, len(sources))
	for _, cmd := range sources {
		name := cmd.Definition().Name
		if _, dup := m[name]; dup {
			return fmt.Errorf("commands: duplicate command name %q", name)
		}
		m[name] = cmd
	}
	mu.Lock()
	live = m
	mu.Unlock()
	return nil
}

func Lookup(name string) (Command, bool) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := live[name]
	return c, ok
}

func Definitions() []*discord.ApplicationCommand {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*discord.ApplicationCommand, 0, len(live))
	for _, c := range live {
		out = append(out, c.Definition())
	}
	return out
}

func Sync(ctx context.Context, c *connection.Client, appID, guild discord.Snowflake) (int, error) {
	defs := Definitions()
	if !guild.IsEmpty() {
		out, err := c.RestClient.BulkOverwriteGuildApplicationCommands(ctx, appID, guild, defs)
		return len(out), err
	}
	out, err := c.BulkRegisterCommands(ctx, defs)
	return len(out), err
}
