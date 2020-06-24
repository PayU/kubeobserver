package common

import (
	"testing"
	"reflect"
)

func TestPodCrashLoopbackStringIdentifier(t *testing.T) {
	s := PodCrashLoopbackStringIdentifier()

	if reflect.TypeOf(s).Kind() != reflect.String || s == "" {
		t.Error("TestPodCrashLoopbackStringIdentifier: couldn't receive PodCrashLoopBack identifier")
	}
}