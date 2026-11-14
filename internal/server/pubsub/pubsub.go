package pubsub

import (
	"sync"
)

type Manager struct {
	// Clients are subscribers. client id's to channel mapping.
	ClientChannels  map[string]map[string]bool
	InSubsciberMode map[string]bool // true if client is in subscribed mode (after SUBSCRIBE command). false otherwise.
	// Client id to Bool mapping.

	// Publishers are channels. Channel to client id's mapping.
	ChannelClients map[string]map[string]bool

	mu sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		ClientChannels:  make(map[string]map[string]bool),
		InSubsciberMode: make(map[string]bool),
		ChannelClients:   make(map[string]map[string]bool),
	}
}

func (m *Manager) SubscribeTo(channel string, clientId string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ClientChannels[clientId] == nil {
		m.ClientChannels[clientId] = make(map[string]bool)
	}
	m.ClientChannels[clientId][channel] = true
	if m.ChannelClients[channel] == nil {
		m.ChannelClients[channel] = make(map[string]bool)
	}
	m.ChannelClients[channel][clientId] = true
	return len(m.ClientChannels[clientId])
}

func (m *Manager) GetChannelSubscribers(channel string) (int, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var subscribers map[string]bool = m.ChannelClients[channel]

	// publish message to all subscribers of the channel
	var clientIds []string
	for clientId := range subscribers {
		clientIds = append(clientIds, clientId)
	}

	return len(subscribers), clientIds
}