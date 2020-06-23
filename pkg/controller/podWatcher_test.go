package controller

import "testing"

func TestNewPodController(t *testing.T) {
	podController := newPodController()

	if podController == nil {
		t.Error("TestNewPodController: couldn't create a new pod controller")
	}
}

func TestPodEventsHandler(t *testing.T) {

}

func TestGetStateChangeOfContainers(t *testing.T) {

}

func TestParseContainerState(t *testing.T) {

}

func TestShouldWatchPod(t *testing.T) {

}

func TestIsSPodControllerSync(t *testing.T) {

}