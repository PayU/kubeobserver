package receivers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PayU/kubeobserver/pkg/common"
	"github.com/PayU/kubeobserver/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

var slackReceiverName = "slack"
var slackAuthorIcon string = "https://raw.githubusercontent.com/kubernetes/community/master/icons/png/resources/unlabeled/pod-128.png"
var slackFooterIcon string = "https://avatars2.githubusercontent.com/u/652790"
var warningIcon string = "https://raw.githubusercontent.com/Keyamoon/IcoMoon-Free/master/PNG/64px/264-warning.png"
var skullIconsSlackStr string = ":skull_and_crossbones::skull_and_crossbones::skull_and_crossbones::skull_and_crossbones::skull_and_crossbones::skull_and_crossbones:"

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
	additionalInfo := receiverEvent.AdditionalInfo
	var colorType string
	var thumbURL string
	var text string

	// this will be true in case some event has slack recevier
	// but no channels were provided in the configuration
	if len(sr.ChannelNames) == 0 {
		c <- errors.New("HandleEvent of slack was triggered but no slack channel names were found in configuration")
		return
	}

	if eventName == "Add" {
		colorType = "good"
	} else if eventName == "Update" {
		colorType = "warning"
	} else if eventName == "Delete" {
		colorType = "danger"
	}

	// no matter what happens, close the channel after function exits
	defer close(c)

	log.Debug().Msg(fmt.Sprintf("received %s message in slack receiver: %s", eventName, message))
	log.Debug().Msg(fmt.Sprintf("building message in Slack format"))

	if additionalInfo[common.PodCrashLoopbackStringIdentifier()] == true { // crash loopback event
		// this will make sure the red color flag
		// add warning thumb on the right side of the meesage.
		// in addition, skull icons will appear on start and the end of the message
		colorType = "#C70039"
		thumbURL = warningIcon
		text = skullIconsSlackStr + message + skullIconsSlackStr
	} else {
		text = "`" + eventName + "`" + " event received: " + message
	}

	attachment := slack.Attachment{
		Color:      colorType,
		AuthorName: "KubeObserver",
		Text:       text,
		Ts:         json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		AuthorIcon: slackAuthorIcon,
		Footer:     "Slack receiver",
		FooterIcon: slackFooterIcon,
		ThumbURL:   thumbURL,
	}

	log.Debug().Msg(fmt.Sprintf("Sending message to Slack: %v", attachment))

	for _, channel := range sr.ChannelNames {
		err := postMessage(sr.SlackClient, channel, &attachment)

		if err != nil {
			var errStr strings.Builder
			errStr.WriteString("slack recevier got unexpected error -> ")
			errStr.WriteString(err.Error())
			c <- errors.New(errStr.String())
		}
	}
}

func postMessage(slackClient *slack.Client, channel string, attachment *slack.Attachment) error {
	channelID, timestamp, err := slackClient.PostMessage(channel, slack.MsgOptionAttachments(*attachment))

	if err == nil {
		log.Debug().Msg(fmt.Sprintf("Succefully posted a message to channel %s at %s", channelID, timestamp))
	} else {
		if strings.HasPrefix(err.Error(), "slack rate limit exceeded") {
			// slack api allows bursts over that limit for short periods. However,
			// if your app continues to exceed its allowance over longer periods of time, we will begin rate limiting.
			// Continuing to send messages after exceeding a rate limit runs the risk of your app being permanently disabled.
			// this this why we are sleeping for 1.5 sec in order to make sure we won't get block
			time.Sleep(1500 * time.Millisecond)
			channelID, timestamp, err = slackClient.PostMessage(channel, slack.MsgOptionAttachments(*attachment))

			if err == nil {
				log.Debug().Msg(fmt.Sprintf("Succefully posted a message to channel %s at %s", channelID, timestamp))
			}
		}
	}

	return err
}
