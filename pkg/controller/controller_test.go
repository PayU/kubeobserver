package controller

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type clientCmd interface {
	buildConfigFromFlags(string, string) (*Config, error)
}

type kuberNetes interface {
	NewForConfig(Config) (ClientSet, error)
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
type mockIndexers map[string]IndexFunc

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

type KeyFunc func(obj interface{}) (string, error)

func KeyFuncImplement(obj interface{}) (string, error) {
	return "mockKey", errors.New("error")
}

type IndexFunc func(obj interface{}) ([]string, error)

func IndexFuncImplement(obj interface{}) ([]string, error) {
	return []string{"mockString1", "mockString2"}, errors.New("error")
}

func TestNewController(t *testing.T) {
	m := make(map[string]IndexFunc)
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	i := cache.NewIndexer(KeyFuncImplement)
}

func TestHomeDir(t *testing.T) {
	result := homeDir()

	if reflect.TypeOf(result).Kind() != reflect.String || result == "" {
		t.Error("TestHomeDir has failed, didn't receive a directory path")
	}
}

func TestInitClientOutOfCluster(t *testing.T) {
	cmd := mockClientCmd{}
	k8s := mockKubernetes{}
	path := "mockPath"
	conf := Config{Master: "mockMaster", Path: &path}

	_, errorFromFlags := cmd.buildConfigFromFlags("mockMaster", &path)

	if errorFromFlags == nil {
		t.Error("error")
	}

	_, errorFromConfig := k8s.NewForConfig(conf)

	if errorFromConfig == nil {
		t.Error("error")
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

}

func TestRun(t *testing.T) {

}

func TestRunWorker(t *testing.T) {

}

func TestSendEventToReceivers(t *testing.T) {

}

func TestWaitForChannelsToClose(t *testing.T) {

}

func TestStartWatch(t *testing.T) {

}
