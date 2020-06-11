package config

import (
	"os"
)

var podEventInDepthDetails bool

func init() {
	podEventInDepthDetails = os.Getenv("IN_DEPTH_POD_DETAILS") == "true"
}

func getInDepth() bool {
	return podEventInDepthDetails
}
