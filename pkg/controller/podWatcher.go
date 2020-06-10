package controller

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

type podContainerStatus struct {
	containerName string
	status        string
}

type podEvent struct {
	EventName   string
	PodName     string
	OldPodState *v1.Pod
}

func printSlice(s []string) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func podUpdateEventHandler(oldPod *v1.Pod, newPod *v1.Pod) {
	fmt.Println("Starting update event")
	oldInitContainerStatuses := oldPod.Status.InitContainerStatuses
	// oldContainerStatuses := oldPod.Status.ContainerStatuses
	newInitContainerStatuses := newPod.Status.InitContainerStatuses
	// newContainerStatuses := newPod.Status.ContainerStatuses

	oldState := make(map[string]string)
	newState := make(map[string]string)

	for _, container := range oldInitContainerStatuses {
		oldState[container.Name] = fmt.Sprintf("The Container %s state %s\n", container.Name, parseContainerState(container.State))
	}

	for _, container := range newInitContainerStatuses {
		newState[container.Name] = fmt.Sprintf("The Container %s state %s\n", container.Name, parseContainerState(container.State))
	}

	for containerName := range newState {
		if oldState[containerName] != newState[containerName] {
			fmt.Println(newState[containerName])
		}
	}

	fmt.Println("finished update event")
}

// podEventsHandler is the business logic of the pod controller.
// In case an error happened, it has to simply return the error.
func podEventsHandler(key string, indexer cache.Indexer) error {
	event := podEvent{}
	json.Unmarshal([]byte(key), &event)

	obj, exists, err := indexer.GetByKey(event.PodName)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", event.PodName, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("Pod %s does not exist anymore\n", event.PodName)
	} else {
		pod := obj.(*v1.Pod)

		switch event.EventName {
		case "Add":
			if (applicationInitTime).Before(pod.ObjectMeta.CreationTimestamp.Time) {
				log.Info().Msg(fmt.Sprintf("Pod %s has been added", pod.ObjectMeta.Name))
			}
		case "Delete":
			log.Info().Msg(fmt.Sprintf("Pod %s has been deleted", pod.ObjectMeta.Name))
		default:
			// update pod evenet
			podUpdateEventHandler(event.OldPodState, pod)
			// formatConditionsArray(pod)
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
			if err == nil {
				out, err := json.Marshal(podEvent{
					EventName:   "Delete",
					PodName:     key,
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

func formatConditionsArray(p *v1.Pod) {

	parseContainerStatus(p.Status.InitContainerStatuses)
	parseContainerStatus(p.Status.ContainerStatuses)
}

func parseContainerStatus(cs []v1.ContainerStatus) {
	for _, s := range cs {
		var containerStateString string = parseContainerState(s.State)
		fmt.Printf("The Container %s state %s", s.Name, containerStateString)

		if s.LastTerminationState.Waiting != nil && s.LastTerminationState.Running != nil && s.LastTerminationState.Terminated != nil {
			parseContainerState(s.LastTerminationState)
			fmt.Printf("%s. Last time this container %s it was after %d restarts", s.Name, containerStateString, s.RestartCount)
		}
	}
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

func parseContainerState2(cs v1.ContainerState) string {
	var s string

	if cs.Waiting != nil {

		s = fmt.Sprint("is waiting ", cs.Waiting.Reason, "\n")

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
