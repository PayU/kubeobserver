package controller

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	asv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var hpaController *controller

type hpaEvent struct {
	EventName  string
	HpaName    string
	NewHpaData *asv1.HorizontalPodAutoscaler
	OldHpaData *asv1.HorizontalPodAutoscaler
}

func newHPAController() *controller {
	// create the hpa watcher
	hpaListWatcher := cache.NewListWatchFromClient(k8sClient.Clientset.AutoscalingV1().RESTClient(), "HorizontalPodAutoscalers", v1.NamespaceAll, fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the hpa key is added to the workqueue.
	indexer, informer := cache.NewIndexerInformer(hpaListWatcher, &asv1.HorizontalPodAutoscaler{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)

			fmt.Println(key)
			fmt.Println(err)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)

			if err == nil {
				out, err := json.Marshal(hpaEvent{
					EventName:  "Update",
					HpaName:    key,
					NewHpaData: new.(*asv1.HorizontalPodAutoscaler),
					OldHpaData: old.(*asv1.HorizontalPodAutoscaler),
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			fmt.Println(key)
			fmt.Println(err)
		},
	}, cache.Indexers{})

	hpaController = newController(queue, indexer, informer, hpaEventsHandler, "HorizontalPodAutoscaler")
	return hpaController
}

// hpaEventsHandler is the business logic of the hpa controller.
// In case an error happened, it has to simply return the error.
func hpaEventsHandler(key string, indexer cache.Indexer) error {
	event := hpaEvent{}
	var eventMessage string
	json.Unmarshal([]byte(key), &event)

	switch event.EventName {
	case "Add":
		log.Debug().Msg(fmt.Sprintf("handling 'Add' event for hpa[%s]", event.HpaName))
	case "Delete":
		log.Debug().Msg(fmt.Sprintf("handling 'Delete' event for hpa[%s]", event.HpaName))
	default:
		// update hpa event
		log.Debug().Msg(fmt.Sprintf("handling 'Update' event for hpa[%s]", event.HpaName))

		var oldHPAStatus, newHPAStatus asv1.HorizontalPodAutoscalerStatus

		if event.OldHpaData != nil && event.NewHpaData != nil {
			oldHPAStatus = event.OldHpaData.Status
			newHPAStatus = event.NewHpaData.Status
		} else {
			log.Warn().Msg(fmt.Sprintf("hpa [%s] old and/or new status is nil. unable to handle 'Update' event", event.EventName))
			return nil
		}

		// new HPA event detected
		if oldHPAStatus.CurrentReplicas == oldHPAStatus.CurrentReplicas {
			if newHPAStatus.CurrentReplicas > newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("scale `DOWN` event has detected by HorizontalPodAutoscaler [`%s`]. starting to `decrease` pod number. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}

			if newHPAStatus.CurrentReplicas < newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("scale `UP` event has detected by HorizontalPodAutoscaler[`%s`]. starting to `increase` pod number.  current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}
		}

		// Scale Up Flow
		if oldHPAStatus.CurrentReplicas < oldHPAStatus.DesiredReplicas {
			if newHPAStatus.CurrentReplicas < newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `UP` event progress has updated. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}

			if newHPAStatus.CurrentReplicas == newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `UP` event has finished. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}
		}

		// Scale Down Flow
		if oldHPAStatus.CurrentReplicas > oldHPAStatus.DesiredReplicas {
			if newHPAStatus.CurrentReplicas > newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `DOWN` event progress has updated. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}

			if newHPAStatus.CurrentReplicas == newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `DOWN` event has finished. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}
		}

		if eventMessage != "" {

		}
	}

	return nil
}
