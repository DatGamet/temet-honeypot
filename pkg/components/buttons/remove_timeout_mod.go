package buttons

import (
	"context"
	"errors"
	"log/slog"
	"temet-honeypot/internal/util"
	"temet-honeypot/pkg/config"
	"temet-honeypot/pkg/database"
	"time"

	"github.com/streame-gg/go-discord-wrapper/api"
	"github.com/streame-gg/go-discord-wrapper/builder"
	"github.com/streame-gg/go-discord-wrapper/connection"
	"github.com/streame-gg/go-discord-wrapper/types/components"
	"github.com/streame-gg/go-discord-wrapper/types/discord"
	"github.com/streame-gg/go-discord-wrapper/types/events"
	"github.com/streame-gg/go-discord-wrapper/types/interactions"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func init() { Register(removeTimeoutMod{}) }

type removeTimeoutMod struct{}

func (removeTimeoutMod) CustomID() string { return "remove_timeout_mod" }

func (removeTimeoutMod) Handle(c *connection.Client, ev *events.InteractionCreateEvent) {
	if !ev.Member.Permissions.Has(discord.PermissionModerateMembers) {
		if _, err := ev.Reply(interactions.ReplyOptions{
			Content: "You do not have permission to use this button.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			slog.Error("failed to reply to interaction", "err", err)
		}
	}

	if database.GlobalConnection == nil {
		if _, err := ev.Reply(interactions.ReplyOptions{
			Content: "An error occurred. Please try again later or contact <@754246997266923571>.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			slog.Error("failed to reply to interaction", "err", err)
		}
		return
	}

	reportCase := database.GlobalConnection.Database().Collection("cases").FindOneAndUpdate(
		context.Background(),
		bson.M{
			"logMessageId": ev.Message.ID,
		},
		map[string]interface{}{
			"$set": map[string]interface{}{
				"resolved":        true,
				"resolvedBy":      ev.Member.UserID,
				"resolveDecision": "REMOVED",
				"resolvedAt":      time.Now(),
			},
		},
	)
	if reportCase.Err() != nil {
		if errors.Is(reportCase.Err(), mongo.ErrNoDocuments) {
			if _, err := ev.Reply(interactions.ReplyOptions{
				Content: "No case found.",
				Flags:   discord.MessageFlagEphemeral,
			}); err != nil {
				slog.Error("failed to reply to interaction", "err", err)
			}
			return
		}
		slog.Error("failed to find and update case", "err", reportCase.Err())
		if _, err := ev.Reply(interactions.ReplyOptions{
			Content: "An error occurred. Please try again later or contact <@754246997266923571>.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			slog.Error("failed to reply to interaction", "err", err)
		}
		return
	}

	var repCase database.Case
	if err := reportCase.Decode(&repCase); err != nil {
		slog.Error("failed to decode case", "err", err)
		if _, err := ev.Reply(interactions.ReplyOptions{
			Content: "An error occurred. Please try again later or contact <@754246997266923571>.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			slog.Error("failed to reply to interaction", "err", err)
		}
		return
	}

	if repCase.Resolved {
		if _, err := ev.Reply(interactions.ReplyOptions{
			Content: "Case is already resolved.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			slog.Error("failed to reply to interaction", "err", err)
		}
		return
	}

	if _, err := c.RestClient.ModifyGuildMember(context.Background(), config.Current.DevGuild, repCase.DiscordUserID, api.ModifyGuildMemberParams{
		CommunicationDisabledUntil: discord.Null[string](),
		AuditLogReason:             util.Pointer("User was timed out due to honeypot, resolved by " + ev.Member.UserID.String()),
	}); err != nil {
		slog.Error("failed to reply to interaction", "err", err)
		if _, err := ev.Reply(interactions.ReplyOptions{
			Content: "An error occurred. Please try again later or contact <@754246997266923571>.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			slog.Error("failed to reply to interaction", "err", err)
		}
		return
	}

	if _, err := ev.Reply(interactions.ReplyOptions{
		Content: "Done. The timeout was removed.",
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		slog.Error("failed to reply to interaction", "err", err)
	}

	logContainer := builder.NewContainer().
		SetAccentColor(0x5865F2).
		AddComponents(
			builder.NewTextDisplay().SetContent("# New User detected\n- Case ID: `"+repCase.MongoID+"`\n- User: <@"+repCase.DiscordUserID.String()+">\n- Resolved: Yes\n- Resolved by: <@"+ev.Member.UserID.String()+">\n- Resolve Decision: Removed timeout").Build(),
			builder.NewSeparator().SetDivider(true).Build(),
			builder.NewActionRow().AddComponents(
				builder.NewButton().
					SetLabel("Keep Timeout").
					SetStyle(components.ButtonStyleSecondary).
					SetCustomID("keep_timeout_mod").
					SetDisabled(true).
					Build(),
				builder.NewButton().
					SetLabel("Remove Timeout").
					SetStyle(components.ButtonStyleSecondary).
					SetCustomID("remove_timeout_mod").
					SetDisabled(true).
					Build(),
				builder.NewButton().
					SetLabel("Ban User").
					SetStyle(components.ButtonStyleDanger).
					SetDisabled(true).
					SetCustomID("ban_user_mod").
					Build(),
			).Build(),
		).
		Build()

	if _, err := c.RestClient.EditMessage(context.Background(), config.Current.HoneypotLogChannel, repCase.LogMessageID, api.EditMessageParams{
		Components: []discord.AnyComponent{logContainer},
	}); err != nil {
		slog.Error("failed to send log message update due to user untimeout", "err", err)
	}
}
