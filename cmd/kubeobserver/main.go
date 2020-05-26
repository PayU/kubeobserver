package main

import (
	"fmt"

	"github.com/shyimo/kubeobserver/pkg/k8sclient"

	"github.com/shyimo/kubeobserver/pkg/handlers"
)

func printAll(vals []interface{}) {
	for _, val := range vals {
		fmt.Println(val)
	}
}

func main() {
	k8sclient.Client()
	handlers.SendMessage("test message for slack")
}
