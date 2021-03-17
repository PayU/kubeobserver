package common

import (
	"strings"

	"github.com/PayU/kubeobserver/pkg/config"
)

var receiversAnnotationName = "kubeobserver.io/receivers"
var podCrashLoopbackStringIdentifier = "CrashLoopBackOff"
var podHpaStringIdentifier = "HorizontalPodAutoscale"

// PodCrashLoopbackStringIdentifier is a getter for k8s api crash loopback string
func PodCrashLoopbackStringIdentifier() string {
	return podCrashLoopbackStringIdentifier
}

// podHpaStringIdentifier is a getter for k8s api hpa
func PodHpaStringIdentifier() string {
	return podHpaStringIdentifier
}

// BuildEventReceiversList builds the list of receivers based on the resource annotations and default configuration
func BuildEventReceiversList(annotations map[string]string) []string {
	eventReceivers := make([]string, 0)

	if annotations != nil && annotations[receiversAnnotationName] != "" {
		eventReceivers = strings.Split(annotations[receiversAnnotationName], ",")
	}

	if len(eventReceivers) == 0 {
		// set the default receiver in case no receiver has mentioned in resource annotation
		eventReceivers = append(eventReceivers, config.DefaultReceiver())
	}

	// remove duplicates if exists

	return eventReceivers
}
