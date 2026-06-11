package components

import (
	"fmt"
	"sync"

	"github.com/streame-gg/go-discord-wrapper/connection"
	"github.com/streame-gg/go-discord-wrapper/types/events"
)

type Handler interface {
	CustomID() string
	Handle(c *connection.Client, ev *events.InteractionCreateEvent)
}

type Registry struct {
	mu      sync.RWMutex
	sources []Handler
	live    map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{live: map[string]Handler{}}
}

func (r *Registry) Register(h Handler) {
	r.sources = append(r.sources, h)
}

func (r *Registry) Lookup(id string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.live[id]
	return h, ok
}

func (r *Registry) Reload() error {
	m := make(map[string]Handler, len(r.sources))
	for _, h := range r.sources {
		id := h.CustomID()
		if _, dup := m[id]; dup {
			return fmt.Errorf("components: duplicate custom ID %q", id)
		}
		m[id] = h
	}
	r.mu.Lock()
	r.live = m
	r.mu.Unlock()
	return nil
}
