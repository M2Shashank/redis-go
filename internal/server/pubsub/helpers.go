package pubsub

func (m *Manager) SubscribeTo(channel string, clientId string) int {
	if m.Subscribers[clientId] == nil {
		m.Subscribers[clientId] = make(map[string]bool)
	}
	m.Subscribers[clientId][channel] = true
	if m.Publishers[channel] == nil {
		m.Publishers[channel] = make(map[string]bool)
	}
	m.Publishers[channel][clientId] = true
	return len(m.Subscribers[clientId])
}

func (m *Manager) PublishMessage(channel string, message string) int {
	var subscribers map[string]bool = m.Publishers[channel]

	// publish message to all subscribers of the channel

	return len(subscribers)
}
