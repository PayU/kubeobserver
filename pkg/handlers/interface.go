package handlers

type messanger interface {
	SendMessage(string, string) error
	GetMessangerType() string
}
