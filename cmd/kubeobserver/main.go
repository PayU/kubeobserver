package main

import (
	"time"

	"github.com/shyimo/kubeobserver/pkg/controller"
	"github.com/shyimo/kubeobserver/pkg/log"
)

func main() {
	log.Initialize()
	controller.StartWatch(time.Now())
}
