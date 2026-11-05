package store

//----- doubly linked list node -----
type ListNode[T any] struct {
	Value T
	Next  *ListNode[T]
	Prev  *ListNode[T]
}

// ------------- doubly linked list implementation --------------
type DoublyLL[T any] struct {
	Head, Tail *ListNode[T]
	Length     int
}

// methods
func (l *DoublyLL[T]) PushFront(value T) {
	var node *ListNode[T] = &ListNode[T]{value, nil, nil}

	if l.Head == nil {
		l.Head = node
		l.Tail = node
	} else {
		node.Next = l.Head
		l.Head.Prev = node
		l.Head = node
	}
	l.Length++
}

func (l *DoublyLL[T]) PushBack(value T) {
	var node *ListNode[T] = &ListNode[T]{value, nil, nil}

	if l.Tail == nil {
		l.Head = node
		l.Tail = node
	} else {
		node.Prev = l.Tail
		l.Tail.Next = node
		l.Tail = node
	}
	l.Length++
}

func (l *DoublyLL[T]) PopFront() (T, bool) {
	if l.Head == nil {
		var zero T
		return zero, false
	}

	value := l.Head.Value
	l.Head = l.Head.Next
	if l.Head == nil {
		l.Tail = nil
	} else {
		l.Head.Prev = nil
	}
	l.Length--

	return value, true
}

func (l *DoublyLL[T]) Front() (T, bool) {
	if l.Head == nil {
		var zero T
		return zero, false
	}
	return l.Head.Value, true
}

func (l *DoublyLL[T]) Back() (T, bool) {
	if l.Tail == nil {
		var zero T
		return zero, false
	}
	return l.Tail.Value, true
}

//for each loop
func (l *DoublyLL[T]) ForEach(f func(value T)) {
	current := l.Head
	for current != nil {
		f(current.Value)
		current = current.Next
	}
}

// ---------- end of DoublyLinkedList Implementation -------------
