package handlers

type messanger interface {
	sendMessage(string, string) error
	getMessangerType() string
}
