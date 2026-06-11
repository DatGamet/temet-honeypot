package modals

import (
	dcomponents "github.com/streame-gg/go-discord-wrapper/types/components"
	"github.com/streame-gg/go-discord-wrapper/types/interactions"
	"github.com/streame-gg/go-discord-wrapper/types/interactions/responses"
	"temet-honeypot/pkg/components"
)

var registry = components.NewRegistry()

func Register(h components.Handler) { registry.Register(h) }

func Lookup(id string) (components.Handler, bool) { return registry.Lookup(id) }

func Reload() error { return registry.Reload() }

func TextValue(i *interactions.Interaction, customID string) string {
	data, ok := interactions.As[*responses.InteractionDataModalSubmit](i.Data)
	if !ok || data.Components == nil {
		return ""
	}
	for _, label := range *data.Components {
		input, ok := label.Component.(*dcomponents.TextInputComponentInteractionResponse)
		if ok && input.CustomID == customID {
			return input.Value
		}
	}
	return ""
}
