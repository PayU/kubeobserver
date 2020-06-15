package controller

import (
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes"
)

func TestHomeDir(t *testing.T) {
	homeDir := homeDir()

	if reflect.TypeOf(homeDir).Kind() != reflect.String || homeDir == "" {
		t.Errorf("Can't get home directory")
	}
}

func TestInitClientOutOfCluster(t *testing.T) {
	var client *kubernetes.Clientset
	client = initClientOutOfCluster()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Falied verifying mandatory Env vars")
		}

		if client == nil {
			t.Errorf("Can't init cient out of cluster")
		}
	}()
}
