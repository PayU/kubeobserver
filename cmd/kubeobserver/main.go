package main

import (
	"time"

	"github.com/shyimo/kubeobserver/pkg/controller"
	"github.com/shyimo/kubeobserver/pkg/log"
	"github.com/shyimo/kubeobserver/pkg/receivers"
)

func main() {
	log.Initialize()
	controller.StartWatch(time.Now())
	receivers.SendMessage("test message for slack")
}
