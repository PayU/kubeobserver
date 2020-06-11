package main

import (
	"time"

	"github.com/shyimo/kubeobserver/pkg/controller"
	"github.com/shyimo/kubeobserver/pkg/handlers"
	"github.com/shyimo/kubeobserver/pkg/log"
)

func main() {
	log.Initialize()
	handlers.SendMessage("test message for slack")
	controller.StartWatch(time.Now())
}
