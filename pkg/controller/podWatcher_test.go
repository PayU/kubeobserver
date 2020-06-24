package controller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestNewPodController(t *testing.T) {
	k8sClient.Clientset = fake.NewSimpleClientset()
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
	cs1 := v1.ContainerStatus{Name: "mockContainer1", Ready: false, RestartCount: 5}
	cs2 := v1.ContainerStatus{Name: "mockContainer2", Ready: true, RestartCount: 0}
	s1 := []v1.ContainerStatus{cs1}
	s2 := []v1.ContainerStatus{cs2}

	results := getStateChangeOfContainers(s1, s2)

	if results == nil {
		t.Error("TestGetStateChangeOfContainers: couldn't parse containers status arrays")
	} else {
		fmt.Println(results)
	}
}

func TestParseContainerState(t *testing.T) {
	csw := v1.ContainerStateWaiting{Reason: "mockWaitingReason", Message: "mockWaitingMessage"}
	cs := v1.ContainerState{Waiting: &csw}
	result := parseContainerState(cs)

	if result == "" {
		t.Error("TestParseContainerStates: couldn't parse container state")
	} else {
		fmt.Println(result)
	}
}

func TestShouldWatchPod(t *testing.T) {
	shouldWatch := shouldWatchPod("mockPod")

	if reflect.TypeOf(shouldWatch).Kind() != reflect.Bool || shouldWatch != true {
		t.Error("TestShouldWatchPod: test hasn't evaluated correctly watch status of mockPod")
	}
}

func TestIsSPodControllerSync(t *testing.T) {
	hasSynced := IsSPodControllerSync()

	if reflect.TypeOf(hasSynced).Kind() != reflect.Bool || hasSynced != false {
		t.Error("TestIsSPodControllerSync: test hasn't evaluated correctly sync status of unexistent controller")
	}
}
