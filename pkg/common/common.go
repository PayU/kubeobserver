package common

import (
	"strings"

	"github.com/PayU/kubeobserver/pkg/config"
)

var receiversAnnotationName = "kubeobserver.io/receivers"
var podCrashLoopbackStringIdentifier = "CrashLoopBackOff"

// PodCrashLoopbackStringIdentifier is a getter for k8s api crash loopback string
func PodCrashLoopbackStringIdentifier() string {
	return podCrashLoopbackStringIdentifier
}

// BuildEventReceiversList builds the list of receivers based on the resource annotations and default configuration
func BuildEventReceiversList(annotations map[string]string) []string {
	eventReceivers := make([]string, 0)

	if annotations != nil && annotations[receiversAnnotationName] != "" {
		eventReceivers = strings.Split(annotations[receiversAnnotationName], ",")
	}

	eventReceivers = append(eventReceivers, config.DefaultReceiver())

	// remove duplicates if exists

	return eventReceivers
}
