package bot

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"temet-honeypot/pkg/database"
	"time"

	"github.com/streame-gg/go-discord-wrapper/cache"
	"github.com/streame-gg/go-discord-wrapper/connection"
	"github.com/streame-gg/go-discord-wrapper/options"
	"github.com/streame-gg/go-discord-wrapper/types/discord"
	devents "github.com/streame-gg/go-discord-wrapper/types/events"

	"temet-honeypot/pkg/commands"
	"temet-honeypot/pkg/components/buttons"
	"temet-honeypot/pkg/components/modals"
	"temet-honeypot/pkg/components/selectmenus"
	"temet-honeypot/pkg/config"
	"temet-honeypot/pkg/events"
)

type Bot struct {
	Client *connection.Client
	db     *database.Connection

	regOnce sync.Once
}

func New(token string, db *database.Connection) (*Bot, error) {
	client, err := connection.NewClient(token,
		discord.IntentGuilds|
			discord.IntentGuildMessages|
			discord.IntentGuildMembers|
			discord.IntentMessageContent,
		options.WithDisableCacheAutoPopulation(),
		options.WithDisableCacheStore(cache.CategoryAll),
		options.WithLogLevel(slog.LevelWarn),
	)
	if err != nil {
		return nil, err
	}

	b := &Bot{Client: client, db: db}

	events.Attach(client)

	for _, reload := range []func() error{commands.Reload, buttons.Reload, selectmenus.Reload, modals.Reload} {
		if err := reload(); err != nil {
			return nil, err
		}
	}

	client.OnReady(func(c *connection.Client, e *devents.ReadyEvent) {
		if err := c.UpdatePresence(connection.UpdatePresenceParams{
			Status: discord.PresenceStatusOnline,
			Activities: []discord.FullActivity{
				{
					Name: "to temet's cord",
					Type: discord.ActivityTypeListening,
				},
			},
		}); err != nil {
			slog.Error("presence update failed", "err", err)
		}

		b.regOnce.Do(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if _, err := commands.Sync(ctx, c, e.Application.ID, config.Current.DevGuild); err != nil {
				c.Logger.Error("command registration failed", "err", err)
			}
		})
	})

	if err := client.OnEvent(devents.EventInteractionCreate, b.route); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bot) route(c *connection.Client, ev *devents.InteractionCreateEvent) {
	switch {
	case ev.IsCommand():
		fields := strings.Fields(ev.GetFullCommand())
		if len(fields) == 0 {
			return
		}
		if cmd, ok := commands.Lookup(fields[0]); ok {
			cmd.Handle(c, ev)
		}
	case ev.IsButton():
		if h, ok := buttons.Lookup(ev.GetCustomID()); ok {
			h.Handle(c, ev)
		}
	case ev.IsAnySelectMenu():
		if h, ok := selectmenus.Lookup(ev.GetCustomID()); ok {
			h.Handle(c, ev)
		}
	case ev.IsModalSubmit():
		if h, ok := modals.Lookup(ev.GetCustomID()); ok {
			h.Handle(c, ev)
		}
	}
}

func (b *Bot) Run(ctx context.Context) error {
	return b.Client.Login(ctx)
}
