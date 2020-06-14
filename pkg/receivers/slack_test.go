package receivers

import (
	"testing"
)

func TestNewSlackReceiver(t *testing.T) {
	dummyURLS := []string{""}
	dummySlackReceiver := NewSlackReceiver(dummyURLS)

	if dummySlackReceiver == nil {
		t.Errorf("Couldn't create a Slack receiver")
	}
}

func TestHandleEvent(t *testing.T) {
	dummyURLS := []string{"https://hooks.slack.com/services/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"}
	dummySlackReceiver := NewSlackReceiver(dummyURLS)
	dummyReceiverEvent := ReceiverEvent{Message: "Dummy message", EventName: "Dummy event"}

	handeled := dummySlackReceiver.HandleEvent(dummyReceiverEvent)

	if handeled != nil {
		t.Errorf("Handling event via a Slack receiver failed")
	}
}
