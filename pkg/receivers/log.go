package receivers

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

var logReceiverName = "log"

// LogReceiver is a struct built for receiving and passing onward events messages to log
type LogReceiver struct{}

func init() {
	ReceiverMap[logReceiverName] = &LogReceiver{}
}

// HandleEvent is an implementation of the Receiver interface for Slack
func (sr *LogReceiver) HandleEvent(receiverEvent ReceiverEvent, c chan error) {
	log.Info().Msg(fmt.Sprintf("log recevier event message[%s]", receiverEvent.Message))
}
