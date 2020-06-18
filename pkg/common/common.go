package common

var podCrashLoopbackStringIdentifier = "CrashLoopBackOff"

// PodCrashLoopbackStringIdentifier is a getter for k8s api crash loopback string
func PodCrashLoopbackStringIdentifier() string {
	return podCrashLoopbackStringIdentifier
}