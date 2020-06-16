package receivers

import (
	"testing"
)

func TestHandleEvent(t *testing.T) {
	dummyNames := []string{"kubeobserver-int-test"}
	dummySlackReceiver := SlackReceiver{ChannelNames: dummyNames}
	dummyReceiverEvent := ReceiverEvent{Message: "Dummy message", EventName: "Dummy event"}
	stopCh := make(chan error)

	dummySlackReceiver.HandleEvent(dummyReceiverEvent, stopCh)

	if e := <-stopCh; e != nil {
		t.Errorf("Falied handling event in slack receiver")
	}
}
