package werks

import (
	"testing"
)

func TestPushMessage(t *testing.T) {
	q := new(MessageQueue)
	q.Capacity = 5
	tests := []string {"one", "two", "three", "four", "five"}
	for _, test := range tests {
		if !q.Push(test) {
			t.Errorf("Unexpected failure in Push")
		}
	}
	if q.Push("Should be maxed.") {
		t.Errorf("Push should have failed, and didn't.")
	}
	if q.Head.Text != "one" {
		t.Errorf("Head should be one and is %s", q.Head.Text)
	}
	if q.Tail.Text != "five" {
		t.Errorf("Tail should be five and is %s", q.Tail.Text)
	}
}

func TestPopMessage(t *testing.T) {
	q := new(MessageQueue)
	q.Capacity = 5
	tests := []string {"one", "two", "three", "four", "five"}
	for _, test := range tests {
		if !q.Push(test) {
			t.Errorf("Unexpected Push failure.")
		}
	}
	for _, test := range tests {
		m := q.Pop()
		if m.Text != test {
			t.Errorf("Got %s, expected %s.", m.Text, test)
		}
	}
	if q.Count != 0 {
		t.Errorf("Queue should be empty, and isn't.")
	}
}
