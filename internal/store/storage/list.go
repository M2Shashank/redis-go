package storage

import (
	"github.com/codecrafters-io/redis-starter-go/internal/collections"
)

type List struct {
	Entries *collections.DoublyLL[string]
	Waiters []chan string
}


// ----------------- List Operations ---------------------

// to be used when dataMutex is already locked
// Used for List operations Only. BLPOP
func (list *List) ServeWaiters() {
	entries, waiters := list.Entries, list.Waiters

	for len(waiters) > 0 && entries.Length > 0 {
		var ch chan string = waiters[0]
		waiters = waiters[1:]

		value, _ := entries.PopFront()
		ch <- value
	}
}
