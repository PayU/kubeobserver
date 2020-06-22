package controller

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/PayU/kubeobserver/pkg/config"
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

func (mc *mockClientCmd) buildConfigFromFlags(masterURL string, kubeconfigPath *string) (*Config, error) {
	conf := Config{Master: masterURL, Path: kubeconfigPath}

	return &conf, errors.New("conf error from flags")
}

func (mk *mockKubernetes) NewForConfig(c Config) (*ClientSet, error) {
	set := ClientSet{Name: c.Master}

	return &set, errors.New("conf error when trying to create a new conf")
}

func (mcont *mockController) processNextItem() bool {
	return true
}

func (mi mockInformer) Run(stopCh <-chan struct{}) {
	return
}

func (mi mockInformer) HasSynced() bool {
	return false
}

func (mi mockInformer) LastSyncResourceVersion() string {
	return "mockVersion"
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

}

func TestRunWorker(t *testing.T) {
	c := mockNewController()
	go c.runWorker()
}

func TestSendEventToReceivers(t *testing.T) {

}

func TestWaitForChannelsToClose(t *testing.T) {

}

func TestStartWatch(t *testing.T) {

}
