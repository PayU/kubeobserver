package receivers

import (
	"testing"
	"time"

	"github.com/slack-go/slack"
)

type slackClient interface {
	HandleEvent(ReceiverEvent, chan error)
	PostMessage(string, slack.MsgOption)
}

type MockSlackClient struct{}

func (m *MockSlackClient) postMessage(ch string, opt slack.MsgOption) (string, time.Time, error) {
	return ch, time.Now(), nil
}

func TestPostMessage(t *testing.T) {
	client := MockSlackClient{}
	now := time.Now()

	ch, time, err := client.postMessage("mockChannel", slack.MsgOptionAsUser(true))

	if ch != "mockChannel" || (time).Before(now) || err != nil {
		t.Error("Error in posting message")
	}
}
