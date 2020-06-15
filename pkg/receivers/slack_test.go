package receivers

import (
	"testing"
)

func TestHandleEvent(t *testing.T) {
	dummyURLS := []string{"https://hooks.slack.com/services/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"}
	dummySlackReceiver := SlackReceiver{ChannelURLS: dummyURLS}
	dummyReceiverEvent := ReceiverEvent{Message: "Dummy message", EventName: "Dummy event"}
	stopCh := make(chan error)

	dummySlackReceiver.HandleEvent(dummyReceiverEvent, stopCh)

	if e := <-stopCh; e != nil {
		t.Errorf("Falied handling event in slack receiver")
	}
}
