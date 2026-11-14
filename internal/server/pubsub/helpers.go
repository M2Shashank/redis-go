package pubsub

func (cm *ClientManager) SubscribeTo(channel string) int {
	cm.subscribedChannels[channel] = true
	return len(cm.subscribedChannels)
}