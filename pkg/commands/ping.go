package commands

import (
	"github.com/streame-gg/go-discord-wrapper/connection"
	dcmd "github.com/streame-gg/go-discord-wrapper/types/commands"
	"github.com/streame-gg/go-discord-wrapper/types/discord"
	"github.com/streame-gg/go-discord-wrapper/types/events"
	"github.com/streame-gg/go-discord-wrapper/types/interactions"
)

func init() { Register(ping{}) }

type ping struct{}

func (ping) Definition() *dcmd.ApplicationCommand {
	return &dcmd.ApplicationCommand{
		Name:        "ping",
		Description: "Check that the bot is alive",
		Type:        discord.ApplicationCommandTypeChatInput,
	}
}

func (ping) Handle(c *connection.Client, ev *events.InteractionCreateEvent) {
	if _, err := ev.Reply(interactions.ReplyOptions{
		Content: "🏓 Pong!",
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		c.Logger.Error("ping reply failed", "err", err)
	}
}
