package receivers

import (
	"errors"
	"testing"
	"time"
	"fmt"

	"github.com/slack-go/slack"
)

type slackClient interface {
	HandleEvent(ReceiverEvent, chan error)
	PostMessage(string, slack.MsgOption)
}

type MockSlackClient struct{}

func (m *MockSlackClient) postMessage(ch string, opt slack.MsgOption) (string, time.Time, error) {
	return mockPostMessage(ch, slack.MsgOptionAsUser(true))
}

func mockPostMessage(ch string, opt slack.MsgOption) (string, time.Time, error) {
	return ch, time.Now(), nil
}

func (m *MockSlackClient) handleEvent(e ReceiverEvent, c chan error) {
	mockHandleEvent(e, c)
}

func mockHandleEvent(e ReceiverEvent, c chan error) {
	text := e.EventName + "" + e.Message

	attach := slack.Attachment{
		Text: text,
	}

	err, _, _ := mockPostMessage("mockChannel", slack.MsgOptionAttachments(attach))

	if err != "mockChannel" {
		c <- errors.New("problem with mockPostMessage")
	}

}

func TestPostMessage(t *testing.T) {
	client := MockSlackClient{}
	now := time.Now()

	ch, time, err := client.postMessage("mockChannel", slack.MsgOptionAsUser(true))

	if ch != "mockChannel" || (time).Before(now) || err != nil {
		t.Error("Error in posting message")
	}
}

func TestHandleEvent(t *testing.T) {
	client := MockSlackClient{}
	event := ReceiverEvent{EventName: "mockEvent", Message: "mockMessage", AdditionalInfo: make(map[string]interface{})}
	channel := make(chan error)

	fmt.Println("starting handle event")
	client.handleEvent(event, channel)
	fmt.Println("finished handle event")

	if err := <-channel; err != nil {
		t.Error("error")
	}
}
