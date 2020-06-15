package config

import (
	"reflect"
	"testing"
)

func TestLogLevel(t *testing.T) {
	logLevel = LogLevel()

	if reflect.TypeOf(logLevel).Kind() != reflect.Int8 || &logLevel == nil {
		t.Errorf("Can't get log level")
	}
}

func TestClusterName(t *testing.T) {
	clusterName := ClusterName()

	if reflect.TypeOf(clusterName).Kind() != reflect.String || clusterName == "" {
		t.Errorf("Can't get cluster name")
	}

}

func TestExcludePodNamePatterns(t *testing.T) {
	excludePodNamePatterns := ExcludePodNamePatterns()

	if reflect.TypeOf(excludePodNamePatterns).Kind() != reflect.Slice {
		t.Errorf("Can't get excluded Pod name patterns")
	}
}

func TestSlackURLS(t *testing.T) {
	slackURLS := SlackURLS()

	if reflect.TypeOf(slackURLS).Kind() != reflect.Slice {
		t.Errorf("Can't get slack URLs")
	}
}

func TestVerifyMandatoryVariables(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Falied verifying mandatory Env vars")
		}
	}()
	verifyMandatoryVariables()
}

func TestSetLogLevel(t *testing.T) {
	setLogLevel()
	logLevel := LogLevel()

	if reflect.TypeOf(logLevel).Kind() != reflect.Int8 || &logLevel == nil {
		t.Errorf("Can't get log level")
	}
}
