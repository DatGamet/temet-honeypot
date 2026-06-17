package events

import (
	"context"
	"log/slog"
	"temet-honeypot/internal"
	"temet-honeypot/internal/util"
	"temet-honeypot/pkg/config"
	"temet-honeypot/pkg/database"
	"time"

	"github.com/streame-gg/go-discord-wrapper/builder"
	"github.com/streame-gg/go-discord-wrapper/connection"
	"github.com/streame-gg/go-discord-wrapper/types/components"
	"github.com/streame-gg/go-discord-wrapper/types/discord"
	devents "github.com/streame-gg/go-discord-wrapper/types/events"
)

var msgMap = internal.NewMessageCache()

func init() {
	On(devents.EventMessageCreate, func(c *connection.Client, e *devents.MessageCreateEvent) {
		if e.Author == nil || !e.GuildID.IsValid() || e.GuildID.IsEmpty() || e.Author.Bot {
			return
		}

		if e.ChannelID != config.Current.HoneypotChannel {
			msgMap.Add(e.Author.ID, e.ChannelID, e.ID)
		}

		if database.GlobalConnection == nil {
			slog.Error("database connection is nil")
			return
		}

		if e.ChannelID == config.Current.HoneypotChannel {
			_, err := c.ModifyGuildMember(context.Background(), *e.GuildID, e.Message.Author.ID,
				discord.MemberEditOptions{
					CommunicationDisabledUntil: util.Pointer(time.Now().Add(24 * time.Hour * 21).Format(time.RFC3339)),
					AuditLogReason:             util.Pointer("detected by sending message in honeypot channel"),
				})
			if err != nil {
				slog.Error("failed to timeout user", "err", err)
			}

			forwardMsg, err := c.CreateMessage(context.Background(), config.Current.HoneypotLogChannel, discord.MessageCreateOptions{
				MessageReference: &discord.MessageMessageReference{
					Type:            util.Pointer(discord.MessageMessageReferenceTypeForward),
					MessageID:       util.Pointer(e.ID),
					ChannelID:       util.Pointer(config.Current.HoneypotChannel),
					GuildID:         util.Pointer(config.Current.DevGuild),
					FailIfNotExists: util.Pointer(true),
				},
			})
			if err != nil {
				slog.Error("failed to forward message", "err", err)
			}

			msgLink := ""
			if forwardMsg != nil {
				msgLink = discord.MessageLink(config.Current.HoneypotLogChannel, forwardMsg.ID, util.Pointer(config.Current.DevGuild))
			}

			logContainer := builder.NewContainer().
				SetAccentColor(0x5865F2).
				AddComponents(
					builder.NewTextDisplay().SetContent("# New User detected\n- Message Reference: "+
						util.InlineIfElse(forwardMsg != nil, msgLink, "Failed to forward message")+
						"\n- Case ID: `"+e.ID.String()+"`\n- User: "+e.Author.Mention()+"\n- Resolved: No\n- Resolved by: None\n- Resolve Decision: None").Build(),
					builder.NewSeparator().SetDivider(true).Build(),
					builder.NewActionRow().AddComponents(
						builder.NewButton().
							SetLabel("Keep Timeout").
							SetStyle(components.ButtonStyleSecondary).
							SetCustomID("keep_timeout_mod").
							Build(),
						builder.NewButton().
							SetLabel("Remove Timeout").
							SetStyle(components.ButtonStyleSecondary).
							SetCustomID("remove_timeout_mod").
							Build(),
					).Build(),
				).
				Build()

			logMessage, err := c.CreateMessage(context.Background(), config.Current.HoneypotLogChannel, discord.MessageCreateOptions{
				Components: []discord.AnyComponent{logContainer},
				Flags:      discord.MessageFlagIsComponentsV2,
			})
			if err != nil {
				slog.Error("failed to send message in log channel", "err", err)
			}

			if err := e.Message.Delete(context.Background(), util.Pointer("Message sent in Honeypot channel")); err != nil {
				slog.Error("failed to delete message in honeypot channel", "err", err)
			}

			if _, err = database.GlobalConnection.Database().Collection("cases").InsertOne(context.Background(), database.NewCase(e.Author.ID, e.ID, logMessage.ID)); err != nil {
				slog.Error("failed to insert case", "err", err)
			}

			dmChannel, err := c.CreateDM(context.Background(), e.Message.Author.ID)
			if err != nil {
				slog.Error("failed to create dm channel with user", "err", err)
			}

			dmContainer := builder.NewContainer().
				SetAccentColor(0x5865F2).
				AddComponents(
					builder.NewTextDisplay().SetContent("# Warning, dear User\n## You sent a message in a honeypot channel.\nWhat is a honeypot?\nA honeypot is a trap designed to punish hackers/scammers. As soon as a user sends a message in this channel, it will be immediately deleted and the user will be punished (ban, timeout, kick).\n\nWhat's the point of this?\nRecently, you can see more and more of the typical 4 pictures scams. Here, hacked accounts send a message in every channel on public Discords, with one or more images, and/or a link to a fake online casino website or other NSFW links / NSFW Discord servers in all channels. This is very annoying and with this channel we want to prevent it.\n\nWhat happens if I send a message here?\nAs soon as you send a message here, it will be immediately deleted and you will be punished. Depending on the server's settings, you will either be banned, timed out, or kicked. The whole point is to prevent this, so it's best not to let it get that far.\n\nHow can I get unbanned?\nEither you can contact a team member directly, or you can simply get unbanned via the button 'Remove Timeout'. In the very last resort, it is also possible to contact <@754246997266923571> via DM.").Build(),
					builder.NewSeparator().SetDivider(true).Build(),
					builder.NewActionRow().AddComponents(
						builder.NewButton().
							SetLabel("Remove Timeout").
							SetCustomID("remove_timeout_user").
							SetStyle(components.ButtonStyleSecondary).
							Build(),
						builder.NewButton().
							SetLabel("How not to get hacked").
							SetStyle(components.ButtonStyleLink).
							SetURL("https://discord.com/safety/360044104071-tips-against-spam-and-hacking").
							Build(),
					).Build(),
				).
				Build()

			//TEMP: wait until library fixes this
			dmChannel.Hydrate(c)

			if _, err := dmChannel.Send(context.Background(), discord.MessageCreateOptions{
				Components: []discord.AnyComponent{dmContainer},
				Flags:      discord.MessageFlagIsComponentsV2,
			}); err != nil {
				slog.Error("failed to send DM channel", "err", err)
			}

			for channelId, messageIds := range msgMap.Get(e.Author.ID) {
				if len(messageIds) == 0 {
					continue
				}

				if len(messageIds) == 1 {
					if err := c.DeleteMessage(context.Background(), channelId, messageIds[0], util.Pointer("Message sent by user that was catched by honeypot")); err != nil {
						slog.Error("failed to delete message in channel", "err", err, "channelId", channelId, "messageId", messageIds[0])
					}
					slog.Info("deleted message in channel", "channelId", channelId, "messageId", messageIds[0])
					continue
				}

				if err := c.BulkDeleteMessages(context.Background(), channelId, messageIds, util.Pointer("Messages sent by user that was catched by honeypot")); err != nil {
					slog.Error("failed to bulk delete message in channel", "err", err, "channelId", channelId, "messages", messageIds)
				}
				slog.Info("deleted messages in channel", "channelId", channelId, "messageId", messageIds)
			}
		}
	})
}
