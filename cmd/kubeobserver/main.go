package main

import (
	"time"

	"github.com/shyimo/kubeobserver/pkg/controller"
	"github.com/shyimo/kubeobserver/pkg/handlers"
)

func main() {
	controller.StartWatch(time.Now())
	handlers.SendMessage("test message for slack")
}
