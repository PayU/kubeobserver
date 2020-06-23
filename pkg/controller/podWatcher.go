package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PayU/kubeobserver/pkg/common"
	"github.com/PayU/kubeobserver/pkg/config"
	"github.com/PayU/kubeobserver/pkg/receivers"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var ignorePodUpdateAnnotationName = "pod-update-kubeobserver.io/ignore"
var receiversAnnotationName = "kubeobserver.io/receivers"
var watchPodInitcontainersAnnotationName = "pod-init-container-kubeobserver.io/watch"
var slackUserIdsAnnotationName = "pod-watch-kubeobserver.io/slack_users_id"

var podController *controller

type podEvent struct {
	EventName  string
	PodName    string
	NewPodData *v1.Pod
	OldPodData *v1.Pod
}

func newPodController() *controller {
	// create the pod watcher
	podListWatcher := cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)

			if err == nil && shouldWatchPod(key) {
				out, err := json.Marshal(podEvent{
					EventName:  "Add",
					PodName:    key,
					NewPodData: obj.(*v1.Pod),
					OldPodData: nil,
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)

			if err == nil && shouldWatchPod(key) {
				out, err := json.Marshal(podEvent{
					EventName:  "Update",
					PodName:    key,
					NewPodData: new.(*v1.Pod),
					OldPodData: old.(*v1.Pod),
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil && shouldWatchPod(key) {
				out, err := json.Marshal(podEvent{
					EventName:  "Delete",
					PodName:    key,
					NewPodData: nil,
					OldPodData: nil,
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
	}, cache.Indexers{})

	podController = newController(queue, indexer, informer, podEventsHandler, "pod")
	return podController
}

// podEventsHandler is the business logic of the pod controller.
// In case an error happened, it has to simply return the error.
func podEventsHandler(key string, indexer cache.Indexer) error {
	event := podEvent{}
	json.Unmarshal([]byte(key), &event)

	podName := event.PodName
	newPod := event.NewPodData
	oldPod := event.OldPodData

	var ignoreEvent bool = false
	var podNamespace string
	var podAnnotations map[string]string
	var podControllerKind string
	var podControllerName string
	var eventMessage strings.Builder
	eventReceivers := make([]string, 0)
	podWatchSlackUsersID := make([]string, 0)

	if newPod != nil {
		podNamespace = newPod.GetNamespace()
		podAnnotations = newPod.GetObjectMeta().GetAnnotations()

		// fetch the pod owner controller
		// this value can be any valid controller like StatefulSet, DaemonSet, ReplicaSet, Job and so on..
		if newPod.GetOwnerReferences() != nil {
			podControllerKind = newPod.GetOwnerReferences()[0].Kind
			podControllerName = newPod.GetOwnerReferences()[0].Name
		}
	}

	if podAnnotations != nil && podAnnotations[receiversAnnotationName] != "" {
		eventReceivers = strings.Split(podAnnotations[receiversAnnotationName], ",")
	}

	eventReceivers = append(eventReceivers, config.DefaultReceiver())

	log.Debug().
		Msg(fmt.Sprintf("found %d event receivers for pod %s in namespace %s. receivers:%s. event-type: %s.",
			len(eventReceivers), podName, podNamespace, strings.Join(eventReceivers, ","), event.EventName))

	switch event.EventName {
	case "Add":
		if (applicationInitTime).Before(newPod.ObjectMeta.CreationTimestamp.Time) {
			messagePodName := podName
			if podControllerKind == "StatefulSet" {
				messagePodName = fmt.Sprintf("%s-%s", podName, newPod.ObjectMeta.UID)
			}

			eventMessage.WriteString(fmt.Sprintf("A `pod` in namesapce `%s` has been `Created`\n", podNamespace))
			eventMessage.WriteString(fmt.Sprintf("Pod name:`%s`\n", messagePodName))
			eventMessage.WriteString(fmt.Sprintf("Environment:`%s`\n", config.ClusterName()))
			eventMessage.WriteString(fmt.Sprintf("Controller kind:`%s`. Controller name:`%s`\n", podControllerKind, podControllerName))
		}

	case "Delete":
		eventMessage.WriteString(fmt.Sprintf("The pod `%s` in `%s` cluster has been deleted\n", podName, config.ClusterName()))
	default:
		// update pod evenet
		watchInitContainers := false
		podUpdates := make([]string, 0)

		// make sure the check update events the happend on the same pod
		if newPod.GetObjectMeta().GetCreationTimestamp() != oldPod.GetObjectMeta().GetCreationTimestamp() {
			return nil
		}

		if podAnnotations != nil {
			ignoreEvent = podAnnotations[ignorePodUpdateAnnotationName] == "true"
			watchInitContainers = podAnnotations[watchPodInitcontainersAnnotationName] == "true"

			if podAnnotations[slackUserIdsAnnotationName] != "" {
				podWatchSlackUsersID = strings.Split(podAnnotations[slackUserIdsAnnotationName], ",")
			}
		}

		if watchInitContainers {
			updates := getStateChangeOfContainers(oldPod.Status.InitContainerStatuses, newPod.Status.InitContainerStatuses)
			podUpdates = append(podUpdates, updates...)
		}

		updates := getStateChangeOfContainers(oldPod.Status.ContainerStatuses, newPod.Status.ContainerStatuses)
		podUpdates = append(podUpdates, updates...)

		if len(podUpdates) > 0 {
			messagePodName := podName
			if podControllerKind == "StatefulSet" {
				messagePodName = fmt.Sprintf("%s-%s", podName, newPod.ObjectMeta.UID)
			}

			eventMessage.WriteString(fmt.Sprintf("A `pod` in namesapce `%s` has been `Updated`. Pod-Name:`%s`. Environment:`%s`.\n", podNamespace, messagePodName, config.ClusterName()))
			eventMessage.WriteString(fmt.Sprintf("Controller kind:`%s`. Controller name:`%s`. Updates:\n", podControllerKind, podControllerName))
			for _, updateStr := range podUpdates {
				eventMessage.WriteString(fmt.Sprintf("- %s", updateStr))
			}
		}
	}

	// if we have any events to update about,
	// send the updates to the relevant receivers
	if eventMessage.String() != "" {
		additionalInfo := make(map[string]interface{})
		onCrashLoopBack := false

		if strings.Contains(eventMessage.String(), common.PodCrashLoopbackStringIdentifier()) {
			additionalInfo[common.PodCrashLoopbackStringIdentifier()] = true
			onCrashLoopBack = true
		}

		additionalInfo["pod_watcher_users_ids"] = podWatchSlackUsersID

		// if we found 'ignore-update-event' annotations but the pod is in crash-loop-back
		// we will still send the event so we can notify about it.
		// in any other case we will send the event as long as 'ignore-update-event' annotations not set to true
		if !ignoreEvent || onCrashLoopBack {
			receiverEvent := receivers.ReceiverEvent{
				EventName:      event.EventName,
				Message:        eventMessage.String(),
				AdditionalInfo: additionalInfo,
			}

			sendEventToReceivers(receiverEvent, eventReceivers)
		}

	}

	return nil
}

// getStateChangeOfContainer will check the different between the continers
// from the old state compare to the new state. the ContainerStatus slice can be the init continers or the reguler continers
// retrun slice of strings that that represents human readable information about the change
func getStateChangeOfContainers(oldContainerStatus []v1.ContainerStatus, newContainerStatus []v1.ContainerStatus) []string {
	result := make([]string, 0)
	oldState := make(map[string]string)
	newState := make(map[string]string)

	for _, container := range oldContainerStatus {
		state := parseContainerState(container.State)
		if state == "" {
			continue
		}

		oldState[container.Name] = fmt.Sprintf("the container %s %s", container.Name, parseContainerState(container.State))
	}

	for _, container := range newContainerStatus {
		state := parseContainerState(container.State)
		if state == "" {
			continue
		}

		newState[container.Name] = fmt.Sprintf("the container %s %s", container.Name, parseContainerState(container.State))
	}

	for containerName := range newState {
		if newState[containerName] != "" && oldState[containerName] != newState[containerName] {
			log.Debug().Msg(fmt.Sprintf("found new state change in %s container. old state:%s. new state:%s.", containerName, oldState[containerName], newState[containerName]))
			result = append(result, newState[containerName])
		}
	}

	return result
}

func parseContainerState(cs v1.ContainerState) string {
	var s string

	if cs.Waiting != nil {
		s = fmt.Sprint("is waiting since ", cs.Waiting.Reason, "\n")

		if cs.Waiting.Message != "" {
			s = fmt.Sprint("is waiting since ", cs.Waiting.Reason, " with following info: ", cs.Waiting.Message, "\n")
		}
	} else if cs.Running != nil {
		s = fmt.Sprint("has started at ", cs.Running.StartedAt, "\n")
	} else if cs.Terminated != nil {
		var reason string

		if cs.Terminated.Reason != "" {
			reason = cs.Terminated.Reason
		} else if cs.Terminated.Message != "" {
			reason = cs.Terminated.Message
		} else {
			return ""
		}

		s = fmt.Sprintf("has been terminated at %v with exit code %d. Reason:`%s`\n", cs.Terminated.FinishedAt, cs.Terminated.ExitCode, reason)
	}

	return s
}

// iterate over the exclude pod name slice
// and check if one (or more) of the slice members contains the pod name
// if so, return false meaning that the event will ignored
func shouldWatchPod(podName string) bool {
	for _, pattern := range config.ExcludePodNamePatterns() {
		if strings.Contains(podName, pattern) {
			return false
		}
	}

	return true
}

// IsSPodControllerSync is used for server health check
func IsSPodControllerSync() bool {
	return podController.informer.HasSynced()
}
