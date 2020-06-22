package controller

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/PayU/kubeobserver/pkg/config"
	"github.com/PayU/kubeobserver/pkg/receivers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type clientCmd interface {
	buildConfigFromFlags(string, string) (*Config, error)
}

type kuberNetes interface {
	NewForConfig(Config) (ClientSet, error)
}

type informer interface {
	Run(stopCh <-chan struct{})
	HasSynced() bool
	LastSyncResourceVersion() string
}

type Config struct {
	Master string
	Path   *string
}
type ClientSet struct {
	Name string
}

type mockClientCmd struct{}
type mockKubernetes struct{}
type mockController struct{}
type mockInformer struct{}
type mockReceiver struct{}

func (mcont *mockController) processNextItem() bool {
	return true
}

func (mi mockInformer) Run(stopCh <-chan struct{}) {
	return
}

func (mi mockInformer) HasSynced() bool {
	return true
}

func (mi mockInformer) LastSyncResourceVersion() string {
	return "mockVersion"
}

func (mr mockReceiver) HandleEvent(r receivers.ReceiverEvent, c chan error) {
	if r.EventName == "Add" {
		fmt.Println("Add event was sent to receiver")
		c <- nil
	} else if r.EventName == "Delete" {
		c <- errors.New("Delete event caused an error")
	} else {
		c <- errors.New("Unexpected error")
	}
}

func KeyFuncImplement(obj interface{}) (string, error) {
	return "mockKey", errors.New("error")
}

func IndexFuncImplement(obj interface{}) ([]string, error) {
	return []string{"mockString1", "mockString2"}, errors.New("error")
}

func mockNewController() *controller {
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	ind := cache.NewIndexer(KeyFuncImplement, cache.Indexers{"mockIndexers": IndexFuncImplement})
	inf := mockInformer{}
	cl := func(str string, i cache.Indexer) error { return errors.New("error") }
	p := "pod"

	return newController(q, ind, inf, cl, p)
}

func TestNewController(t *testing.T) {
	mockController := mockNewController()

	if mockController == nil {
		t.Error("error")
	}
}

func TestHomeDir(t *testing.T) {
	result := homeDir()

	if reflect.TypeOf(result).Kind() != reflect.String || result == "" {
		t.Error("TestHomeDir has failed, didn't receive a directory path")
	}
}

func TestInitClientOutOfCluster(t *testing.T) {
	var kubeconfig *string = config.KubeConfFilePath()
	client := initClientOutOfCluster()

	if _, err := os.Stat(*kubeconfig); os.IsNotExist(err) && client != nil {
		t.Error("Though config file doesn't exist, somehow a k8s client was initiated")
	}
}

func TestProcessNextItem(t *testing.T) {
	controller := mockController{}

	processed := controller.processNextItem()

	if !processed {
		t.Error("error")
	}
}

func TestHandleErr(t *testing.T) {
	c := mockNewController()
	err := errors.New("mockError")
	key := "mockKey"

	c.queue.AddRateLimited(key)
	c.handleErr(nil, key)
	newKey, _ := c.queue.Get()

	if newKey != key {
		t.Error("error")
	}

	c.queue.Done(key)
	c.handleErr(err, key)
	newKey, _ = c.queue.Get()

	if newKey != key {
		t.Error("error")
	}
}

func TestRun(t *testing.T) {
	mockController := mockNewController()
	threads := 1
	c := make(chan struct{})
	key := "mockKey"

	mockController.queue.AddRateLimited(key)

	go mockController.Run(threads, c)

	mockController.queue.Forget(key)
	mockController.queue.Done(key)
	close(c)
}

func TestRunWorker(t *testing.T) {
	c := mockNewController()
	go c.runWorker()
}

func TestSendEventToReceivers(t *testing.T) {
	addEvent := receivers.ReceiverEvent{EventName: "Add", Message: "mockMessage", AdditionalInfo: make(map[string]interface{})}
	deleteEvent := receivers.ReceiverEvent{EventName: "Delete", Message: "mockMessage", AdditionalInfo: make(map[string]interface{})}
	receiversSlice := []string{"mockReceiver"}
	r := mockReceiver{}
	receivers.ReceiverMap[receiversSlice[0]] = r

	sendEventToReceivers(addEvent, receiversSlice)
	sendEventToReceivers(deleteEvent, receiversSlice)
}

func TestWaitForChannelsToClose(t *testing.T) {
	var channelList []chan error
	channel := make(chan error)
	channelList = append(channelList, channel)

	go waitForChannelsToClose(channelList...)

	for _, c := range channelList {
		c <- errors.New("mockError")
		close(c)
	}
}

func TestStartWatch(t *testing.T) {
	go StartWatch(time.Now())
}
