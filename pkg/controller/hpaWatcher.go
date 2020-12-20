package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PayU/kubeobserver/pkg/common"
	"github.com/PayU/kubeobserver/pkg/config"
	"github.com/PayU/kubeobserver/pkg/receivers"
	"github.com/rs/zerolog/log"
	"k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const hpaSlackUserIdsAnnotationName = "hpa-watch-kubeobserver.io/slack_users_id"

var hpaController *controller

type hpaEvent struct {
	EventName  receivers.EventName
	HpaName    string
	NewHpaData *v2beta1.HorizontalPodAutoscaler
	OldHpaData *v2beta1.HorizontalPodAutoscaler
}

func newHPAController() *controller {
	// create the hpa watcher
	hpaListWatcher := cache.NewListWatchFromClient(k8sClient.Clientset.AutoscalingV2beta1().RESTClient(), "HorizontalPodAutoscalers", v1.NamespaceAll, fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the hpa key is added to the workqueue.
	indexer, informer := cache.NewIndexerInformer(hpaListWatcher, &v2beta1.HorizontalPodAutoscaler{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)

			if err == nil {
				out, err := json.Marshal(hpaEvent{
					EventName:  receivers.AddEvent,
					HpaName:    key,
					NewHpaData: obj.(*v2beta1.HorizontalPodAutoscaler),
					OldHpaData: nil,
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)

			log.Info().Msg("in HPA update event")

			if err == nil {
				out, err := json.Marshal(hpaEvent{
					EventName:  receivers.UpdateEvent,
					HpaName:    key,
					NewHpaData: new.(*v2beta1.HorizontalPodAutoscaler),
					OldHpaData: old.(*v2beta1.HorizontalPodAutoscaler),
				})

				if err == nil {
					log.Info().Msg("second error is nil")
					queue.Add(string(out))
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)

			if err == nil {
				out, err := json.Marshal(hpaEvent{
					EventName:  receivers.DeleteEvent,
					HpaName:    key,
					NewHpaData: nil,
					OldHpaData: obj.(*v2beta1.HorizontalPodAutoscaler),
				})

				if err == nil {
					queue.Add(string(out))
				}
			}
		},
	}, cache.Indexers{})

	hpaController = newController(queue, indexer, informer, hpaEventsHandler, "HorizontalPodAutoscaler")
	return hpaController
}

// hpaEventsHandler is the business logic of the hpa controller.
// In case an error happened, it has to simply return the error.
func hpaEventsHandler(key string, indexer cache.Indexer) error {
	log.Info().Msg("Starting hpaEventsHandler")
	event := hpaEvent{}
	json.Unmarshal([]byte(key), &event)

	var eventMessage string
	var hpaAnnotations map[string]string
	hpaWatchSlackUsersID := make([]string, 0)

	if event.NewHpaData != nil {
		hpaAnnotations = event.NewHpaData.GetObjectMeta().GetAnnotations()
	}

	eventReceivers := common.BuildEventReceiversList(hpaAnnotations)
	log.Debug().Msg(fmt.Sprintf("found %d event receivers for HorizontalPodAutoscaler[%s]. receivers[%s]", len(eventReceivers), event.HpaName, strings.Join(eventReceivers, ",")))

	switch event.EventName {
	case "Add":
		log.Debug().Msg(fmt.Sprintf("handling 'Add' event for HorizontalPodAutoscaler[%s]", event.HpaName))
		eventMessage = fmt.Sprintf("New HorizontalPodAutoscaler resource [`%s`] added to `%s` cluster", event.HpaName, config.ClusterName())
		log.Debug().Msg(eventMessage)

	case "Delete":
		log.Debug().Msg(fmt.Sprintf("handling 'Delete' event for HorizontalPodAutoscaler[%s]", event.HpaName))
		eventMessage = fmt.Sprintf("HorizontalPodAutoscaler resource [`%s`] has deleted from `%s` cluster", event.HpaName, config.ClusterName())
		log.Debug().Msg(eventMessage)

	default:
		// update hpa event
		log.Debug().Msg(fmt.Sprintf("handling 'Update' event for HorizontalPodAutoscaler[%s]", event.HpaName))

		var oldHPAStatus, newHPAStatus v2beta1.HorizontalPodAutoscalerStatus

		if event.OldHpaData != nil && event.NewHpaData != nil {
			oldHPAStatus = event.OldHpaData.Status
			newHPAStatus = event.NewHpaData.Status
		} else {
			log.Warn().Msg(fmt.Sprintf("HorizontalPodAutoscaler [%s] old and/or new status is nil. unable to handle 'Update' event", event.EventName))
			return nil
		}

		fmt.Printf("%v", oldHPAStatus)
		fmt.Println("------------------")
		fmt.Printf("%v", newHPAStatus)

		// new HPA event detected
		if oldHPAStatus.CurrentReplicas == oldHPAStatus.CurrentReplicas {
			if newHPAStatus.CurrentReplicas > newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("scale `DOWN` event has detected by HorizontalPodAutoscaler [`%s`] in `%s` cluster. starting to `decrease` pod number. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, config.ClusterName(), newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}

			if newHPAStatus.CurrentReplicas < newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("scale `UP` event has detected by HorizontalPodAutoscaler[`%s`] in `%s` cluster. starting to `increase` pod number.  current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, config.ClusterName(), newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}
		}

		// Scale Up Flow
		if oldHPAStatus.CurrentReplicas < oldHPAStatus.DesiredReplicas {
			if newHPAStatus.CurrentReplicas < newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `UP` event progress has updated in `%s` cluster. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, config.ClusterName(), newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}

			if newHPAStatus.CurrentReplicas == newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `UP` event has finished in `%s` cluster. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, config.ClusterName(), newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}
		}

		// Scale Down Flow
		if oldHPAStatus.CurrentReplicas > oldHPAStatus.DesiredReplicas {
			if newHPAStatus.CurrentReplicas > newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `DOWN` event progress has updated in `%s` cluster. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, config.ClusterName(), newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}

			if newHPAStatus.CurrentReplicas == newHPAStatus.DesiredReplicas {
				eventMessage = fmt.Sprintf("HorizontalPodAutoscaler[`%s`] scale `DOWN` event has finished in `%s` cluster. current-replicas:`%d` desired-replicas:`%d`",
					event.HpaName, config.ClusterName(), newHPAStatus.CurrentReplicas, newHPAStatus.DesiredReplicas)
				log.Debug().Msg(eventMessage)
			}
		}

		if hpaAnnotations != nil && hpaAnnotations[hpaSlackUserIdsAnnotationName] != "" {
			hpaWatchSlackUsersID = strings.Split(hpaAnnotations[hpaSlackUserIdsAnnotationName], ",")
		}
	}

	if eventMessage != "" {
		additionalInfo := make(map[string]interface{})
		additionalInfo["pod_watcher_users_ids"] = hpaWatchSlackUsersID

		receiverEvent := receivers.ReceiverEvent{
			EventName:      event.EventName,
			Message:        eventMessage,
			AdditionalInfo: additionalInfo,
		}

		sendEventToReceivers(receiverEvent, eventReceivers)
	}

	return nil
}
