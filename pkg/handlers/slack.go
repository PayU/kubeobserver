package handlers

import "fmt"

// SendMessage sending a message
func SendMessage(msg string) {
	fmt.Println("Sending message to slack -> ", msg)
}
