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
	Prev  *Element
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
func (q *Queue) PushMessage(text string) {
	m := new(TextMessage)
	m.Text = text
	q.Push(m)
}

// Push pushes a new Element onto the head of the queue.  If the
// queue is at capacity, it pops the tail to make room.
func (q *Queue) Push(value interface{}) {
	for q.Count >= q.Capacity {
		q.Pop()
	}

	e := new(Element)
	e.Value = value
	e.Prev = nil

	if q.Count <= 0 {
		q.Head = e
		q.Tail = e
		e.Next = nil
	} else {
		e.Next = q.Head
		q.Head.Prev = e
		q.Head = e
	}
	q.Count += 1
}

// Pop pops the tail of the queue.
func (q *Queue) Pop() *Element {
	if q.Count <= 0 {
		return nil
	}
	e := q.Tail
	if q.Count == 1 {
		q.Head = nil
		q.Tail = nil
	} else {
		q.Tail = e.Prev
		q.Tail.Next = nil
	}
	q.Count -= 1

	e.Prev = nil
	e.Next = nil
	return e
}

// Peek returns the tail of the queue without popping it.
func (q *Queue) Peek() *Element {
	if q.Count <= 0 {
		return nil
	}
	return q.Tail
}

// PopMessage pops a text message from the tail of the queue.
func (q *Queue) PopMessage() string {
	e := q.Pop()
	if e == nil {
		return ""
	}

	return e.Value.(*TextMessage).Text
}
