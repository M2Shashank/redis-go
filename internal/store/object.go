package store

type ObjectType int

const StringType ObjectType = 0
const ListType ObjectType = 1

type RedisObject struct {
	Type ObjectType
	Value any
}


//----- doubly linked list node for list implementation -----
type DoublyLL struct {
	Head, Tail *ListNode
	Length int
}

type ListNode struct {
	Value string
	Next  *ListNode
	Prev  *ListNode
}

func (l *DoublyLL) PushFront(value string) {
	var node *ListNode = &ListNode{value, nil, nil}

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

func (l *DoublyLL) PushBack(value string) {
	var node *ListNode = &ListNode{value, nil, nil}

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

func (l *DoublyLL) PopFront() (string, bool) {
	if l.Head == nil {
		return "", false
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

// ---------------- end of List Implementation -----------------