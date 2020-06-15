package receivers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shyimo/kubeobserver/pkg/config"
	"github.com/slack-go/slack"
)

var slackReceiverName = "slack"

// SlackReceiver is a struct built for receiving and passing onward events messages to Slack
type SlackReceiver struct {
	ChannelURLS []string
}

func init() {
	ReceiverMap[slackReceiverName] = &SlackReceiver{
		ChannelURLS: config.SlackURLS(),
	}
}

// HandleEvent is
func (sr *SlackReceiver) HandleEvent(receiverEvent ReceiverEvent, c chan error) {
	chanelURLS := sr.ChannelURLS
	message := receiverEvent.Message
	eventName := receiverEvent.EventName

	log.Info().Msg(fmt.Sprintf("Received %s message in slack receiver: %s", eventName, message))
	log.Info().Msg(fmt.Sprintf("Building message in Slack format"))

	attachment := slack.Attachment{
		Color:      "good",
		AuthorName: "kubeobserver",
		Text:       "`" + eventName + "`" + " event received: " + message,
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	log.Info().Msg(fmt.Sprintf("Sending message to Slack: "))

	for _, url := range chanelURLS {
		defer close(c)
		innerURL := url
		go func() {
			err := slack.PostWebhook(innerURL, &msg)

			if err != nil {
				log.Error().Msg(fmt.Sprintf("Got error while posting webhook to Slack: %s", err))
				c <- err
			}
		}()
	}
}

// NewSlackReceiver create new slack receiverz
func NewSlackReceiver(channelURLS []string) *SlackReceiver {
	return &SlackReceiver{
		ChannelURLS: channelURLS,
	}
}
