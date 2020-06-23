package controller

import (
	"encoding/json"
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func TestNewPodController(t *testing.T) {
	podController := newPodController()

	if podController == nil {
		t.Error("TestNewPodController: couldn't create a new pod controller")
	}
}

func TestPodEventsHandler(t *testing.T) {
	i := cache.NewIndexer(KeyFuncImplement, cache.Indexers{"mockIndexers": IndexFuncImplement})
	mockPod1 := &v1.Pod{}
	mockPod2 := &v1.Pod{}
	mockPodEvent := podEvent{EventName: "mockPodEvent", PodName: "mockPod", OldPodData: mockPod1, NewPodData: mockPod2}
	mockPodEventAsString, strErr := json.Marshal(mockPodEvent)

	if strErr == nil {
		err := podEventsHandler(string(mockPodEventAsString), i)

		if err != nil {
			t.Error("TestPodEventsHandler: couldn't handle pod event properly")
		}
	} else {
		fmt.Println("TestPodEventsHandler: couldn't marshal podEvent to byte array, then to string")
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("TestPodEventsHandler: test failed and recovered since: %s \n", r)
		}
	}()
}

func TestGetStateChangeOfContainers(t *testing.T) {

}

func TestParseContainerState(t *testing.T) {

}

func TestShouldWatchPod(t *testing.T) {

}

func TestIsSPodControllerSync(t *testing.T) {

}
