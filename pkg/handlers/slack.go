package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	slack "github.com/shyimo/kubeobserver/pkg/handlers/slack"
	"github.com/slack-go/slack"
)

type slackMessanger struct {
	SendingFunc func()
	Type        string
}

// NewSlackMessanger create new slack messanger
func NewSlackMessanger() {
	return &slackMessanger{SendingFunc: slack.sendMessage(), Type: "slack"}
}

// SendMessage sending a message
func (s slackMessanger) sendMessage(message string, url string) {
	// api := slack.New("xoxb-3128660797-1178532843059-l9PE1P4l8wDaiVsLBfC0KItW")
	// fmt.Println("Sending message to slack -> ", msg)

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
	err := slack.PostWebhook("https://hooks.slack.com/services/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", &msg)
	if err != nil {
		fmt.Println(err)
	}
}

func getMessangerType() string {
	return "slack"
}
