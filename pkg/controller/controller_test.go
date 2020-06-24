package controller

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/PayU/kubeobserver/pkg/config"
	"github.com/PayU/kubeobserver/pkg/receivers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type informer interface {
	Run(stopCh <-chan struct{})
	HasSynced() bool
	LastSyncResourceVersion() string
}

type mockInformer struct{}
type mockReceiver struct{}

func (mi mockInformer) Run(stopCh <-chan struct{}) {
	return
}

func (mi mockInformer) HasSynced() bool {
	return true
}

func (mi mockInformer) LastSyncResourceVersion() string {
	return "mockInformerVersion"
}

func (mr mockReceiver) HandleEvent(r receivers.ReceiverEvent, c chan error) {
	if r.EventName == "Add" {
		fmt.Println("Add event was sent to mockReceiver")
		c <- nil
	} else if r.EventName == "Delete" {
		c <- errors.New("Delete event caused an error in mockReceiver")
	} else {
		c <- errors.New("Unexpected error in mockReceiver")
	}
}

func KeyFuncImplement(obj interface{}) (string, error) {
	return "mockKey", errors.New("mockError")
}

func IndexFuncImplement(obj interface{}) ([]string, error) {
	return []string{"mockString1", "mockString2"}, errors.New("mockEerror")
}

func mockNewController() *controller {
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	ind := cache.NewIndexer(KeyFuncImplement, cache.Indexers{"mockIndexers": IndexFuncImplement})
	inf := mockInformer{}
	cl := func(str string, i cache.Indexer) error { return errors.New("mockErrorInControllerLogic") }
	p := "pod"

	return newController(q, ind, inf, cl, p)
}

func TestNewController(t *testing.T) {
	mockController := mockNewController()

	if mockController == nil {
		t.Error("TestNewController: error in initializing a new controller")
	}
}

func TestHomeDir(t *testing.T) {
	result := homeDir()

	if reflect.TypeOf(result).Kind() != reflect.String || result == "" {
		t.Error("TestHomeDir: test has failed, didn't receive a directory path")
	}
}

func TestInitClientOutOfCluster(t *testing.T) {
	var kubeconfig *string = config.KubeConfFilePath()
	client := initClientOutOfCluster()

	if _, err := os.Stat(*kubeconfig); os.IsNotExist(err) && client != nil {
		t.Error("TestInitClientOutOfCluster: Though config file doesn't exist, somehow a k8s client was initiated")
	}
}

func TestProcessNextItem(t *testing.T) {
	controller := mockNewController()
	key := "mockKeyFromProcessNextItem"
	var processed bool

	controller.queue.AddRateLimited(key)
	processed = controller.processNextItem()

	if !processed {
		t.Error("TestProcessNextItem: test has failed")
	}
}

func TestHandleErr(t *testing.T) {
	c := mockNewController()
	err := errors.New("mockError")
	key := "mockKeyFromHandleErr"

	c.queue.AddRateLimited(key)
	c.handleErr(nil, key)
	newKey, _ := c.queue.Get()

	if newKey != key {
		t.Error("TestHandleErr: key wasn't properly managed in controller")
	}

	c.queue.Done(key)
	c.handleErr(err, key)
	newKey, _ = c.queue.Get()

	if newKey != key {
		t.Error("TestHandleErr: error wasn't properly managed in controller")
	}
}

func TestRun(t *testing.T) {
	mockController := mockNewController()
	threads := 1
	c := make(chan struct{})
	key := "mockKeyFromRun"

	mockController.queue.AddRateLimited(key)

	go mockController.Run(threads, c)

	mockController.queue.Forget(key)
	mockController.queue.Done(key)

	defer func() {
		close(c)
		_, ok := <-c

		if ok {
			t.Error("TestRun: channel wasn't closed properly")
		} else if r := recover(); r != nil {
			t.Errorf("TestRun: unexpectedly failed with error: %s \n", r)
		}
	}()
}

func TestRunWorker(t *testing.T) {
	c := mockNewController()
	go c.runWorker()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestRunWorker: unexpectedly failed with error: %s \n", r)
		}
	}()
}

func TestSendEventToReceivers(t *testing.T) {
	addEvent := receivers.ReceiverEvent{EventName: "Add", Message: "mockMessage", AdditionalInfo: make(map[string]interface{})}
	deleteEvent := receivers.ReceiverEvent{EventName: "Delete", Message: "mockMessage", AdditionalInfo: make(map[string]interface{})}
	receiversSlice := []string{"mockReceiver"}
	r := mockReceiver{}
	receivers.ReceiverMap[receiversSlice[0]] = r

	sendEventToReceivers(addEvent, receiversSlice)
	sendEventToReceivers(deleteEvent, receiversSlice)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestSendEventToReceivers: unexpectedly failed with error: %s \n", r)
		}
	}()
}

func TestWaitForChannelsToClose(t *testing.T) {
	var channelList []chan error
	channel := make(chan error)
	channelList = append(channelList, channel)

	go waitForChannelsToClose(channelList...)

	for _, c := range channelList {
		c <- errors.New("mockErrorFromWaitForChannelsToClose")
		close(c)
		_, ok := <-c

		if ok {
			t.Error("TestWaitForChannelsToClose: channel wasn't closed properly")
		}
	}
}
