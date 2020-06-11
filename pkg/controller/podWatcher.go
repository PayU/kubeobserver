package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shyimo/kubeobserver/pkg/config"
	handlers "github.com/shyimo/kubeobserver/pkg/handlers"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type podEvent struct {
	EventName   string
	PodName     string
	OldPodState *v1.Pod
}

// getStateChangeOfContainer will check the different between the continers
// from the old state compare to the new state. the ContainerStatus slice can be the init continers or the reguler continers
// return true in case the state changed and the information about the change
func getStateChangeOfContainer(oldContainerStatus []v1.ContainerStatus, newContainerStatus []v1.ContainerStatus) (bool, []string) {
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

	return len(result) > 0, result
}

// podEventsHandler is the business logic of the pod controller.
// In case an error happened, it has to simply return the error.
func podEventsHandler(key string, indexer cache.Indexer) error {
	event := podEvent{}
	json.Unmarshal([]byte(key), &event)
	slackMessanger := handlers.NewSlackMessanger("https://hooks.slack.com/services/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")

	obj, exists, err := indexer.GetByKey(event.PodName)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Fetching object with key %s from store failed with %v", event.PodName, err))
		return err
	}

	if !exists {
		log.Info().Msg(fmt.Sprintf("got empty result from controller indexer while trying to fetch %s pod", event.PodName))
	} else {
		pod := obj.(*v1.Pod)
		var eventMessage strings.Builder

		switch event.EventName {
		case "Add":
			if (applicationInitTime).Before(pod.ObjectMeta.CreationTimestamp.Time) {
				log.Info().Msg(fmt.Sprintf("Pod %s has been added", pod.ObjectMeta.Name))
				slackMessanger.SendMessage(fmt.Sprintf("Pod %s has been added", pod.ObjectMeta.Name), slackMessanger.ChannelURL)
			}
		case "Delete":
			eventMessage.WriteString(fmt.Sprintf("the pod %s in %s cluster has been deleted", pod.ObjectMeta.Name, config.ClusterName()))
			slackMessanger.SendMessage(fmt.Sprintf("the pod %s in %s cluster has been deleted", pod.ObjectMeta.Name, config.ClusterName()), slackMessanger.ChannelURL)
		default:
			// update pod evenet
			fmt.Println("Starting update event")
			// oldContainerStatuses := oldPod.Status.ContainerStatuses

			// newContainerStatuses := newPod.Status.ContainerStatuses

			// pod init containers status change check
			isStateChange, updates := getStateChangeOfContainer(event.OldPodState.Status.InitContainerStatuses, pod.Status.InitContainerStatuses)
			if isStateChange {
				fmt.Println(updates)
				slackMessanger.SendMessage(strings.Join(updates, " "), slackMessanger.ChannelURL)
			}

			// pod main containers status change check
			// isStateChange, updates = getStateChangeOfContainer(event.OldPodState.Status.ContainerStatuses, pod.Status.ContainerStatuses)
			// if isStateChange {
			// 	fmt.Println(updates)
			// }

			fmt.Println("finished update event")
			slackMessanger.SendMessage("finished update event", slackMessanger.ChannelURL)
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
