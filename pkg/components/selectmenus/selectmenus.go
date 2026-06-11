package selectmenus

import (
	"temet-honeypot/pkg/components"

	"github.com/streame-gg/go-discord-wrapper/types/interactions"
	"github.com/streame-gg/go-discord-wrapper/types/interactions/responses"
)

var registry = components.NewRegistry()

func Register(h components.Handler) { registry.Register(h) }

func Lookup(id string) (components.Handler, bool) { return registry.Lookup(id) }

func Reload() error { return registry.Reload() }

func SelectedValues(i *interactions.Interaction) []string {
	if data, ok := interactions.As[*responses.InteractionDataMessageComponent](i.Data); ok {
		return data.Values
	}
	return nil
}
