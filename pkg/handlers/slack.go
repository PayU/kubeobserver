package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

// SlackMessanger send messages to slack
type SlackMessanger struct {
	Type       string
	ChannelURL string
}

// NewSlackMessanger create new slack messanger
func NewSlackMessanger(channelURL string) *SlackMessanger {
	return &SlackMessanger{
		Type:       "slack",
		ChannelURL: channelURL,
	}
}

// SendMessage sending a message
func (s SlackMessanger) SendMessage(message string, url string) error {
	log.Info().Msg(fmt.Sprintf("Received message in slack messanger: %s \n", message))

	attachment := slack.Attachment{
		Color:      "good",
		AuthorName: "kubeobserver",
		Text:       message,
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}

	log.Info().Msg(fmt.Sprintf("sending message to slack"))
	err := slack.PostWebhook(url, &msg)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s SlackMessanger) GetMessangerType() string {
	return "slack"
}
