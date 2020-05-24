package main

import (
	"fmt"

	"github.com/shyimo/kubeobserver/pkg/handlers"
)

func printAll(vals []interface{}) {
	for _, val := range vals {
		fmt.Println(val)
	}
}

func main() {
	handlers.SendMessage("test message for slack")

	

	var names = []interface{"shai", "coral", "itay"}
	printAll(names)
}
