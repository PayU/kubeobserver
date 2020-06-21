package receivers

import (
	"errors"
	"testing"
	"time"

	"github.com/slack-go/slack"
)

type slackClient interface {
	postMessage(string, slack.MsgOption)
}

type MockSlackClient struct{}
type MockSlackReceiver struct{}

func (m *MockSlackClient) postMessage(ch string, opt slack.MsgOption) (string, time.Time, error) {
	return ch, time.Now(), nil
}

func (mr *MockSlackReceiver) postMessage(mc *MockSlackClient, channel string, attachement *slack.Attachment) error {
	_, _, err := mc.postMessage(channel, slack.MsgOptionAttachments(*attachement))

	return err
}

func (mr *MockSlackReceiver) handleEvent(e ReceiverEvent, c chan error) {
	client := MockSlackClient{}
	text := e.EventName + "" + e.Message

	attach := slack.Attachment{
		Text: text,
	}

	err := mr.postMessage(&client, "mockChannel", &attach)

	if err != nil {
		c <- errors.New("problem with mockPostMessage")
	}
}

func TestPostMessage(t *testing.T) {
	reciever := MockSlackReceiver{}
	client := MockSlackClient{}
	attach := slack.Attachment{}

	err := reciever.postMessage(&client, "mockChannel", &attach)

	if err != nil {
		t.Error(err)
	}
}

func TestHandleEvent(t *testing.T) {
	reciever := MockSlackReceiver{}
	event := ReceiverEvent{EventName: "mockEvent", Message: "mockMessage", AdditionalInfo: make(map[string]interface{})}
	channel := make(chan error)

	reciever.handleEvent(event, channel)

	select {
	case err := <-channel:
		t.Error(err)
	default:
	}
}
