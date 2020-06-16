package controller

import (
	"reflect"
	"testing"
)

func TestNewPodController(t *testing.T) {
	podController := newPodController()

	if podController == nil {
		t.Errorf("Can't create a pod controller")
	}
}

func TestShouldWatchPod(t *testing.T) {
	shouldWatch := shouldWatchPod("dummyName")

	if reflect.TypeOf(shouldWatch).Kind() != reflect.Bool {
		t.Errorf("Can't understand weather to watch Pod or not")
	}
}
