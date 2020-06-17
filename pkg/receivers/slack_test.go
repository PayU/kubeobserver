package receivers

import (
	"testing"

	"github.com/shyimo/kubeobserver/pkg/config"
	"github.com/slack-go/slack"
)

func TestHandleEvent(t *testing.T) {
	dummyNames := config.SlackChannelNames()
	dummySlackReceiver := SlackReceiver{ChannelNames: dummyNames, SlackClient: slack.New(config.SlackToken())}
	dummyReceiverEvent := ReceiverEvent{Message: "Dummy message", EventName: "Dummy event"}
	stopCh := make(chan error)

	dummySlackReceiver.HandleEvent(dummyReceiverEvent, stopCh)

	if e := <-stopCh; e != nil {
		t.Errorf("Falied handling event in slack receiver")
	}
}
