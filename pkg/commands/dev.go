package commands

import (
	"context"
	"fmt"
	"temet-honeypot/internal/util"
	"temet-honeypot/pkg/components/buttons"
	"time"

	"github.com/streame-gg/go-discord-wrapper/connection"
	"github.com/streame-gg/go-discord-wrapper/types/discord"
	"github.com/streame-gg/go-discord-wrapper/types/events"
	"github.com/streame-gg/go-discord-wrapper/types/interactions"

	"temet-honeypot/pkg/components/modals"
	"temet-honeypot/pkg/components/selectmenus"
	"temet-honeypot/pkg/config"
	tevents "temet-honeypot/pkg/events"
)

func init() { Register(dev{}) }

type dev struct{}

func (dev) Definition() *discord.ApplicationCommand {
	return &discord.ApplicationCommand{
		Name:                     "dev",
		Description:              "Owner-only developer tools",
		Type:                     discord.ApplicationCommandTypeChatInput,
		DefaultMemberPermissions: util.Pointer(discord.PermissionAdministrator),
		Options: []discord.AnyApplicationCommandOption{
			&discord.ApplicationCommandOptionSubCommand{
				Type:        discord.ApplicationCommandOptionTypeSubCommand,
				Name:        "reload",
				Description: "Reload commands, events, or components",
				Options: []discord.AnyApplicationCommandOption{
					&discord.ApplicationCommandOptionString{
						Type:        discord.ApplicationCommandOptionTypeString,
						Name:        "choice",
						Description: "What to reload",
						Required:    true,
						Choices: []discord.ApplicationCommandOptionChoice[string]{
							{Name: "Commands", Value: "commands"},
							{Name: "Events", Value: "events"},
							{Name: "Components", Value: "components"},
							{Name: "All", Value: "all"},
						},
					},
				},
			},
		},
	}
}

func (dev) Handle(c *connection.Client, ev *events.InteractionCreateEvent) {
	if !config.Current.IsOwner(ev.Member.UserID) {
		_, _ = ev.Reply(interactions.ReplyOptions{
			Content: "⛔ This command is owner-only.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if ev.GetSubCommand() != "reload" {
		return
	}

	choice := ev.GetStringOption("choice")
	msg, err := reload(c, &ev.Interaction, choice)
	if err != nil {
		msg = "⚠️ " + err.Error()
	}
	_, _ = ev.Reply(interactions.ReplyOptions{Content: msg, Flags: discord.MessageFlagEphemeral})
}

func reload(c *connection.Client, i *interactions.Interaction, choice string) (string, error) {
	switch choice {
	case "events":
		tevents.Reload()
		return "🔁 Reloaded event handlers.", nil
	case "components":
		if err := reloadComponents(); err != nil {
			return "", fmt.Errorf("component reload failed: %w", err)
		}
		return "🔁 Reloaded button, select-menu, and modal handlers.", nil
	case "commands":
		if err := Reload(); err != nil {
			return "", fmt.Errorf("command reload failed: %w", err)
		}
		n, err := syncCommands(c, i.ApplicationID)
		if err != nil {
			return "", fmt.Errorf("command sync failed: %w", err)
		}
		return fmt.Sprintf("🔁 Reloaded command handlers and re-registered %d commands.", n), nil
	case "all":
		tevents.Reload()
		if err := reloadComponents(); err != nil {
			return "", fmt.Errorf("component reload failed: %w", err)
		}
		if err := Reload(); err != nil {
			return "", fmt.Errorf("reloaded events and components, but command reload failed: %w", err)
		}
		n, err := syncCommands(c, i.ApplicationID)
		if err != nil {
			return "", fmt.Errorf("reloaded events and components, but command sync failed: %w", err)
		}
		return fmt.Sprintf("🔁 Reloaded events, components, and command handlers, and re-registered %d commands.", n), nil
	default:
		return "", fmt.Errorf("unknown reload choice %q", choice)
	}
}

// reloadComponents rebuilds every component registry: buttons, select menus,
// and modals. It returns the first duplicate-custom-ID error so the caller can
// report it; a failed reload leaves that registry's previous table intact.
func reloadComponents() error {
	for _, reload := range []func() error{buttons.Reload, selectmenus.Reload, modals.Reload} {
		if err := reload(); err != nil {
			return err
		}
	}
	return nil
}

// syncCommands re-registers the command set with Discord, scoped to the
// configured dev guild when set. It shares commands.Sync with startup so the
// two never disagree about where commands live.
func syncCommands(c *connection.Client, appID discord.Snowflake) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return Sync(ctx, c, appID, config.Current.DevGuild)
}
