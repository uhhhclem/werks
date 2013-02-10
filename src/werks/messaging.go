package werks

// TextMessage represents a simple text message.
type TextMessage struct {
	Text string `json:"text"`
}

// ChatMessage represents a chat message from a player.  It's not
// JSON serializable.
type ChatMessage struct {
	Player *Player
	Text   string
}

// Element is an element in a queue.
type Element struct {
	Next  *Element
	Value interface{}
}

// Queue represents a queue of Elements.
type Queue struct {
	Head     *Element
	Tail     *Element
	Count    int
	Capacity int
}

// PushMessage pushes a text message onto the head of the queue.
func (q *Queue) PushMessage(text string) bool {
	m := new(TextMessage)
	m.Text = text
	return q.Push(m)
}

func (q *Queue) Push(value interface{}) bool {
	if q.Count == q.Capacity {
		return false
	}

	e := new(Element)
	e.Value = value
	e.Next = nil

	if q.Count <= 0 {
		q.Head = e
		q.Tail = e
	} else {
		q.Tail.Next = e
		q.Tail = e
	}
	q.Count += 1
	return true
}

func (q *Queue) Pop() *Element {
	if q.Count <= 0 {
		return nil
	}
	e := q.Head
	if q.Count == 1 {
		q.Head = nil
		q.Tail = nil
	} else {
		q.Head = q.Head.Next
	}
	q.Count -= 1
	return e

}

// PopMessage pops a text message from the tail of the queue.
func (q *Queue) PopMessage() string {
	e := q.Pop()
	if e == nil {
		return ""
	}

	return e.Value.(*TextMessage).Text
}
