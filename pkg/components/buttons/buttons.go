package buttons

import "temet-honeypot/pkg/components"

var registry = components.NewRegistry()

func Register(h components.Handler) { registry.Register(h) }

func Lookup(id string) (components.Handler, bool) { return registry.Lookup(id) }

func Reload() error { return registry.Reload() }
