package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var mandatoryEnvironmentVariables = []string{"K8S_CLUSTER_NAME"}

var k8sClusterName string
var logLevel zerolog.Level
var excludePodNamePatterns []string
var slackURLS []string
var defaultReceiver string

func init() {
	setLogLevel()
	verifyMandatoryVariables()

	k8sClusterName = os.Getenv("K8S_CLUSTER_NAME")

	if os.Getenv("SLACK_URLS") == "" {
		slackURLS = make([]string, 0)
	} else {
		slackURLS = strings.Split(os.Getenv("SLACK_URLS"), ",")
	}

	if os.Getenv("EXCLUDE_POD_NAME_PATTERNS") == "" {
		excludePodNamePatterns = make([]string, 0)
	} else {
		excludePodNamePatterns = strings.Split(os.Getenv("EXCLUDE_POD_NAME_PATTERNS"), ",")
	}

	if defaultReceiver = os.Getenv("DEFAULT_RECEIVER"); defaultReceiver == "" {
		defaultReceiver = "slack"
	}

	outputConfig()
}

// LogLevel is a getter for zerolog log level
func LogLevel() zerolog.Level {
	return logLevel
}

// ClusterName is a getter function for the current running cluster
func ClusterName() string {
	return k8sClusterName
}

// ExcludePodNamePatterns is a getter function for the excludePodNamePatterns slice
func ExcludePodNamePatterns() []string {
	return excludePodNamePatterns
}

// SlackURLS is a getter funcrtion for the slackURLS slice
func SlackURLS() []string {
	return slackURLS
}

// DefaultReceiver is a getter function for DefaultReceiver string
func DefaultReceiver() string {
	return defaultReceiver
}

func verifyMandatoryVariables() {
	for _, envVar := range mandatoryEnvironmentVariables {
		if os.Getenv(envVar) == "" {
			panic(fmt.Sprintf("missing mandatory environment variable: %s", envVar))
		}
	}
}

func setLogLevel() {
	logLevelStr := os.Getenv("LOG_LEVEL")

	if logLevelStr == "" {
		logLevel = zerolog.InfoLevel
		return
	}

	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	}
}

func outputConfig() {
	log.Info().
		Str("k8sClusterName", k8sClusterName).
		Str("logLevel", logLevel.String()).
		Str("excludePodNamePatterns", strings.Join(excludePodNamePatterns, " ")).
		Str("defaultReceiver", defaultReceiver).
		Msg("kubeobserver configurations")
}
