package werks

import (
	"testing"
)

func TestPushMessage(t *testing.T) {
	var text string
	q := new(Queue)
	q.Capacity = 5
	tests := []string{"one", "two", "three", "four", "five"}
	for _, test := range tests {
		if !q.PushMessage(test) {
			t.Errorf("Unexpected failure in Push")
		}
	}
	if q.PushMessage("Should be maxed.") {
		t.Errorf("PushMessage should have failed, and didn't.")
	}
	text = q.Head.Value.(*TextMessage).Text
	if text != "one" {
		t.Errorf("Head should be one and is %s", text)
	}
	text = q.Tail.Value.(*TextMessage).Text
	if text != "five" {
		t.Errorf("Tail should be five and is %s", text)
	}
}

func TestPopMessage(t *testing.T) {
	q := new(Queue)
	q.Capacity = 5
	tests := []string{"one", "two", "three", "four", "five"}
	for _, test := range tests {
		if !q.PushMessage(test) {
			t.Errorf("Unexpected PushMessage failure.")
		}
	}
	for _, test := range tests {
		s := q.PopMessage()
		if s != test {
			t.Errorf("Got %s, expected %s.", s, test)
		}
	}
	if q.Count != 0 {
		t.Errorf("Queue should be empty, and isn't.")
	}
}
