package receivers

import (
	"fmt"
	"time"
)

var slackReceiverName = "slack"

type slackReceiver struct{}

func init() {
	ReceiverMap[slackReceiverName] = &slackReceiver{}
}

// HandleEvent is
func (sr *slackReceiver) HandleEvent(receiverEvent ReceiverEvent, c chan error) {
	defer close(c)

	fmt.Println(receiverEvent.Message)
	time.Sleep(5000 * time.Millisecond)
}
