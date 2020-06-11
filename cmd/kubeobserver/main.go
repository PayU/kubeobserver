package main

import (
	"time"

	"github.com/shyimo/kubeobserver/pkg/controller"
	"github.com/shyimo/kubeobserver/pkg/handlers"
)

func main() {
	handlers.SendMessage("test message for slack")
	controller.StartWatch(time.Now())
}
