// Package config holds the template's compile-time settings.
//
// Unlike the bot token (read from the environment in main.go), these are values
// a developer edits directly in source: who may run privileged commands, and
// which guild to register commands in while developing. Edit Current below.
package config

import "github.com/streame-gg/go-discord-wrapper/types/discord"

type Config struct {
	OwnerID            discord.Snowflake
	DevGuild           discord.Snowflake
	HoneypotChannel    discord.Snowflake
	HoneypotLogChannel discord.Snowflake
}

var Current = Config{
	OwnerID:            754246997266923571,
	DevGuild:           897183099731976192,
	HoneypotChannel:    1509274129201889351,
	HoneypotLogChannel: 1509274217579937974,
}

func (c Config) IsOwner(id discord.Snowflake) bool {
	return c.OwnerID != 0 && id == c.OwnerID
}
