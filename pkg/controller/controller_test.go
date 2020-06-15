package controller

import (
	"fmt"
	"reflect"
	"testing"
)

func TestHomeDir(t *testing.T) {
	fmt.Println("checking home dir")
	homeDir := homeDir()

	if reflect.TypeOf(homeDir).Kind() != reflect.String || homeDir == "" {
		fmt.Println("error in home dir")
		t.Errorf("Can't get home directory")
	}
}

// func TestInitClientOutOfCluster(t *testing.T) {
// 	var client *kubernetes.Clientset
// 	client = initClientOutOfCluster()

// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Errorf("Falied verifying mandatory Env vars")
// 		}

// 		if client == nil {
// 			t.Errorf("Can't init cient out of cluster")
// 		}
// 	}()
// }
