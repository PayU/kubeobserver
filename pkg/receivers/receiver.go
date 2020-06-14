package receivers

// ReceiverMap is 
var ReceiverMap = make(map[string]Receiver) 

// Receiver interace 
type Receiver interface{
	HandleEvent(receiverEvent ReceiverEvent) error
}

// ReceiverEvent is bla bla bla
type ReceiverEvent struct {
	EventName string
	Message string
}