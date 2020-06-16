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
	ChannelNames []string
}

func init() {
	ReceiverMap[slackReceiverName] = &SlackReceiver{
		ChannelNames: config.SlackChannelNames(),
	}
}

// HandleEvent is an implementation of the Receiver interface for Slack
func (sr *SlackReceiver) HandleEvent(receiverEvent ReceiverEvent, c chan error) {
	chanelNames := sr.ChannelNames
	slackToken := config.SlackToken()
	message := receiverEvent.Message
	eventName := receiverEvent.EventName
	authorIcon := "https://raw.githubusercontent.com/kubernetes/community/master/icons/png/resources/unlabeled/pod-128.png"
	footerIcon := "https://avatars2.githubusercontent.com/u/652790"
	var colorType string

	if eventName == "Add" {
		colorType = "good"
	} else if eventName == "Update" {
		colorType = "warning"
	} else if eventName == "Delete" {
		colorType = "danger"
	}

	// no matter what happens, close the channel after function exits
	defer close(c)

	log.Debug().Msg(fmt.Sprintf("Received %s message in slack receiver: %s", eventName, message))
	log.Debug().Msg(fmt.Sprintf("Building message in Slack format"))

	slackAPI := slack.New(slackToken)

	attachment := slack.Attachment{
		Color:      colorType,
		AuthorName: "KubeObserver",
		Text:       "`" + eventName + "`" + " event received: " + message,
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		AuthorIcon: authorIcon,
		Footer:     "Slack receiver",
		FooterIcon: footerIcon,
	}

	log.Debug().Msg(fmt.Sprintf("Sending message to Slack: %v", attachment))

	for _, channel := range chanelNames {
		channelID, timestamp, err := slackAPI.PostMessage(channel, slack.MsgOptionAttachments(attachment))

		if err == nil {
			log.Debug().Msg(fmt.Sprintf("Succefully posted a message to channel %s at %s", channelID, timestamp))
		} else {
			log.Error().Msg(fmt.Sprintf("An error occured when posting a message to slack: %v", err))
			c <- err
		}
	}
}
