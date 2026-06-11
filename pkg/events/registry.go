package events

import (
	"sync"

	"github.com/streame-gg/go-discord-wrapper/connection"
	devents "github.com/streame-gg/go-discord-wrapper/types/events"
)

type Handler func(*connection.Client, devents.Event)

type registration struct {
	event   devents.EventType
	handler Handler
}

var registrations []registration

func On[T devents.Event](event devents.EventType, h func(*connection.Client, T)) {
	registrations = append(registrations, registration{
		event: event,
		handler: func(c *connection.Client, ev devents.Event) {
			if typed, ok := ev.(T); ok {
				h(c, typed)
			}
		},
	})
}

var (
	mu       sync.RWMutex
	live     map[devents.EventType][]Handler
	attached = map[devents.EventType]bool{}
)

func Attach(c *connection.Client) {
	rebuild()

	mu.RLock()
	types := make([]devents.EventType, 0, len(live))
	for et := range live {
		types = append(types, et)
	}
	mu.RUnlock()

	for _, et := range types {
		if attached[et] {
			continue
		}
		attached[et] = true

		et := et
		_ = c.OnEvent(et, connection.EventHandler(func(c *connection.Client, ev devents.Event) {
			mu.RLock()
			hs := live[et]
			mu.RUnlock()
			for _, h := range hs {
				h(c, ev)
			}
		}))
	}
}

func Reload() { rebuild() }

func rebuild() {
	m := make(map[devents.EventType][]Handler)
	for _, r := range registrations {
		m[r.event] = append(m[r.event], r.handler)
	}
	mu.Lock()
	live = m
	mu.Unlock()
}
