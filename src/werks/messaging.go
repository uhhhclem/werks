package werks

// Message represents a message that needs to be sent to the front
// end.
type Message struct {
	Text string `json:"text"`
	Next *Message `json:"-"`
}

// MessageQueue represents a queue of messages that are to be sent to
// the front end.
type MessageQueue struct {
	Head *Message
	Tail *Message
	Count int
	Capacity int
}

func (q *MessageQueue) Push(text string) bool {
	m := new(Message)
	m.Text = text
	m.Next = nil

	if q.Count == q.Capacity {
		return false
	}

	if q.Count <= 0 {
		q.Head = m
		q.Tail = m
	} else {
		q.Tail.Next = m
		q.Tail = m
	}
	q.Count += 1
	return true
}

func (q *MessageQueue) Pop() *Message {
	if q.Count <= 0 {
		return nil
	}
	m := q.Head
	if q.Count == 1 {
		q.Head = nil
		q.Tail = nil
	} else {
		q.Head = q.Head.Next
	}
	q.Count -= 1
	return m
}
