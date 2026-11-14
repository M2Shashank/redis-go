package pubsub

func (m *Manager) SubscribeTo(channel string, clientId string) int {
	m.Subscribers[clientId][channel] = true
	m.Publishers[channel][clientId] = true
	return len(m.Subscribers[clientId])
}

func (m *Manager) PublishMessage(channel string, message string) int {
	var subscribers map[string]bool = m.Publishers[channel]

	// publish message to all subscribers of the channel

	return len(subscribers)
}