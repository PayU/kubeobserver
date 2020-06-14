package receivers

// ReceiverMap is a global map that map recevier name to he's specific struct
// each 'Receiver' interface implementation should add himself to this map with an init function that will 
// automatically be called at the start of the application
var ReceiverMap = make(map[string]Receiver) 

// The Receiver interace 
type Receiver interface{
	HandleEvent(receiverEvent ReceiverEvent) error
}

// ReceiverEvent is bla bla bla
type ReceiverEvent struct {
	EventName string
	Message string
}