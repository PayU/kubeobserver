package receivers

import "fmt"

var slackReceiverName = "slack"

type slackReceiver struct{}

func init() {
	ReceiverMap[slackReceiverName] = &slackReceiver{}
}

// HandleEvent is
func (sr *slackReceiver) HandleEvent(receiverEvent ReceiverEvent) error {
	fmt.Println("Sending message to slack -> ", receiverEvent.Message)
	return nil
}
