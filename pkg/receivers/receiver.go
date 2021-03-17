package receivers

type EventName string

const (
	// AddEvent is bla
	AddEvent EventName = "Add"
	// DeleteEvent is bla
	DeleteEvent EventName = "EventName"
	// UpdateEvent is bla
	UpdateEvent EventName = "Update"
)

// ReceiverMap is a global map that map receiver name to he's specific struct
// each 'Receiver' interface implementation should add himself to this map with an init function that will
// automatically be called at the start of the application
var ReceiverMap = make(map[string]Receiver)

// The Receiver interface
type Receiver interface {
	HandleEvent(receiverEvent ReceiverEvent, c chan error)
}

// ReceiverEvent represent any processed event
// from a watcher (pod watcher, config-map watcher and so on..)
type ReceiverEvent struct {
	EventName      EventName
	Message        string
	AdditionalInfo map[string]interface{}
}
