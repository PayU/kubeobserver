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
var slackAuthorIcon string = "https://raw.githubusercontent.com/kubernetes/community/master/icons/png/resources/unlabeled/pod-128.png"
var slackFooterIcon string = "https://avatars2.githubusercontent.com/u/652790"

// SlackReceiver is a struct built for receiving and passing onward events messages to Slack
type SlackReceiver struct {
	ChannelNames []string
	SlackClient  *slack.Client
}

func init() {
	ReceiverMap[slackReceiverName] = &SlackReceiver{
		ChannelNames: config.SlackChannelNames(),
		SlackClient:  slack.New(config.SlackToken()),
	}
}

// HandleEvent is an implementation of the Receiver interface for Slack
func (sr *SlackReceiver) HandleEvent(receiverEvent ReceiverEvent, c chan error) {
	message := receiverEvent.Message
	eventName := receiverEvent.EventName
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

	attachment := slack.Attachment{
		Color:      colorType,
		AuthorName: "KubeObserver",
		Text:       "`" + eventName + "`" + " event received: " + message,
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		AuthorIcon: slackAuthorIcon,
		Footer:     "Slack receiver",
		FooterIcon: slackFooterIcon,
	}

	log.Debug().Msg(fmt.Sprintf("Sending message to Slack: %v", attachment))

	for _, channel := range sr.ChannelNames {
		channelID, timestamp, err := sr.SlackClient.PostMessage(channel, slack.MsgOptionAttachments(attachment))

		if err == nil {
			log.Debug().Msg(fmt.Sprintf("Succefully posted a message to channel %s at %s", channelID, timestamp))
		} else {
			log.Error().Msg(fmt.Sprintf("An error occured when posting a message to slack: %v", err))
			c <- err
		}
	}
}
