package main

import (
	"github.com/fatih/color"
	"github.com/shyimo/kubeobserver/pkg/handlers"
	"rsc.io/quote"
)

func main() {
	handlers.SendMessage()
	color.Cyan(quote.Hello())
}
