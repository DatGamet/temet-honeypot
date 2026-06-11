package modals

import (
	"github.com/streame-gg/go-discord-wrapper/connection"
	"github.com/streame-gg/go-discord-wrapper/types/discord"
	"github.com/streame-gg/go-discord-wrapper/types/events"
	"github.com/streame-gg/go-discord-wrapper/types/interactions"
)

func init() { Register(feedback{}) }

type feedback struct{}

func (feedback) CustomID() string { return "demo_feedback" }

func (feedback) Handle(c *connection.Client, ev *events.InteractionCreateEvent) {
	msg := TextValue(&ev.Interaction, "message")
	if _, err := ev.Reply(interactions.ReplyOptions{
		Content: "📝 Thanks for the feedback:\n> " + msg,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		c.Logger.Error("feedback modal reply failed", "err", err)
	}
}
