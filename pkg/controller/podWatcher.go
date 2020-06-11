package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shyimo/kubeobserver/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var watchPodInitcontainersAnnotationName = "init-container-kubeobserver.io/watch"

type podEvent struct {
	EventName   string
	PodName     string
	OldPodState *v1.Pod
}

// getStateChangeOfContainer will check the different between the continers
// from the old state compare to the new state. the ContainerStatus slice can be the init continers or the reguler continers
// retrun slice of strings that that represents human readable information about the change
func getStateChangeOfContainers(oldContainerStatus []v1.ContainerStatus, newContainerStatus []v1.ContainerStatus) []string {
	result := make([]string, 0)
	oldState := make(map[string]string)
	newState := make(map[string]string)

	for _, container := range oldContainerStatus {
		oldState[container.Name] = fmt.Sprintf("the container %s %s\n", container.Name, parseContainerState(container.State))
	}

	for _, container := range newContainerStatus {
		newState[container.Name] = fmt.Sprintf("the container %s %s\n", container.Name, parseContainerState(container.State))
	}

	for containerName := range newState {
		if oldState[containerName] != newState[containerName] {
			result = append(result, newState[containerName])
		}
	}

	return result
}

// podEventsHandler is the business logic of the pod controller.
// In case an error happened, it has to simply return the error.
func podEventsHandler(key string, indexer cache.Indexer) error {
	event := podEvent{}
	json.Unmarshal([]byte(key), &event)

	obj, exists, err := indexer.GetByKey(event.PodName)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("fetching object with key %s from store failed with %v", event.PodName, err))
		return err
	}

	if !exists {
		log.Info().Msg(fmt.Sprintf("got empty result from controller indexer while trying to fetch %s pod", event.PodName))
	} else {
		pod := obj.(*v1.Pod)
		podName := pod.ObjectMeta.Name
		var eventMessage strings.Builder
		podAnnotations := pod.GetObjectMeta().GetAnnotations()

		switch event.EventName {
		case "Add":
			if (applicationInitTime).Before(pod.ObjectMeta.CreationTimestamp.Time) {
				eventMessage.WriteString(fmt.Sprintf("the pod %s has been added to %s cluster", podName, config.ClusterName()))
			}
		case "Delete":
			eventMessage.WriteString(fmt.Sprintf("the pod %s in %s cluster has been deleted", podName, config.ClusterName()))
		default:
			// update pod evenet
			watchInitContainers := false
			podUpdates := make([]string, 0)

			if podAnnotations != nil {
				watchInitContainers = podAnnotations[watchPodInitcontainersAnnotationName] == "true"
			}

			if watchInitContainers {
				updates := getStateChangeOfContainers(event.OldPodState.Status.InitContainerStatuses, pod.Status.InitContainerStatuses)
				podUpdates = append(podUpdates, updates...)
			}

			updates := getStateChangeOfContainers(event.OldPodState.Status.ContainerStatuses, pod.Status.ContainerStatuses)
			podUpdates = append(podUpdates, updates...)

			if len(podUpdates) > 0 {
				fmt.Println("new update has arrived in pod", podName)
				fmt.Println(podUpdates)
			}
		}
	}

	return nil
}

func newPodController() *controller {
	// create the pod watcher
	podListWatcher := cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(), "pods", "logs", fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				out, err := json.Marshal(podEvent{
					EventName:   "Add",
					PodName:     key,
					OldPodState: nil,
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				out, err := json.Marshal(podEvent{
					EventName:   "Update",
					PodName:     key,
					OldPodState: old.(*v1.Pod),
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			fmt.Println(obj.(cache.DeletedFinalStateUnknown).Key)
			fmt.Println(obj.(cache.DeletedFinalStateUnknown).Obj.(*v1.Pod).UID)

			podName := key
			if err == nil {
				out, err := json.Marshal(podEvent{
					EventName:   "Delete",
					PodName:     podName,
					OldPodState: nil,
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
	}, cache.Indexers{})

	controller := newController(queue, indexer, informer, podEventsHandler, "pod")

	// We can now warm up the cache for initial synchronization.
	// Let's suppose that we knew about a pod "testPod" on our last run, therefore add it to the cache.
	// If this pod is not there anymore, the controller will be notified about the removal after the
	// cache has synchronized.
	indexer.Add(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testPod",
			Namespace: v1.NamespaceDefault,
		},
	})

	return controller
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
		s = fmt.Sprint("has been terminated at ", cs.Terminated.FinishedAt, " with status code ", cs.Terminated.ExitCode, " after receiving a ", cs.Terminated.Signal, " signal \n")
	}

	return s
}
