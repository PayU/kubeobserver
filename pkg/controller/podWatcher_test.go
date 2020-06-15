package controller

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewPodController(t *testing.T) {
	fmt.Println("checking pod controller")
	podController := newPodController()

	if podController == nil {
		fmt.Println("error in pod controller")
		t.Errorf("Can't create a pod controller")
	}
}

func TestShouldWatchPod(t *testing.T) {
	fmt.Println("checking should watch pod")
	shouldWatch := shouldWatchPod("dummyName")

	if reflect.TypeOf(shouldWatch).Kind() != reflect.Bool {
		fmt.Println("error in should watch pod")
		t.Errorf("Can't understand weather to watch Pod or not")
	}
}
