package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PayU/kubeobserver/pkg/config"
	"github.com/PayU/kubeobserver/pkg/receivers"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

type k8sClientStruct struct {
	Clientset kubernetes.Interface
}

var k8sClient k8sClientStruct
var applicationInitTime time.Time

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h // linux
	}

	return os.Getenv("USERPROFILE") // windows
}

func initClientOutOfCluster() *kubernetes.Clientset {
	var kubeconfig *string = config.KubeConfFilePath()

	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(r.(string), "stat") {
				log.Error().Msg("Local config file probably doesn't exist in default path")
			} else {
				panic(errors.New(r.(string)))
			}
		}
	}()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// controllerLogic types take an string & cache.Indexer. it return an error value (if occuer).
// this will be used by the controller struct
type controllerLogic func(string, cache.Indexer) error

type controller struct {
	indexer      cache.Indexer
	queue        workqueue.RateLimitingInterface
	informer     cache.Controller
	eventHandler controllerLogic
	resourceType string
}

func newController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, handler controllerLogic, resourceType string) *controller {
	return &controller{
		informer:     informer,
		indexer:      indexer,
		queue:        queue,
		eventHandler: handler,
		resourceType: resourceType,
	}
}

func (c *controller) processNextItem() bool {
	log.Debug().Msg(fmt.Sprintf("waiting for new items to appear on the queue of %s controller", c.resourceType))
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.eventHandler(key.(string), c.indexer)
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		log.Error().Msg(fmt.Sprintf("Error syncing pod %v: %v", key, err))

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Info().Msg(fmt.Sprintf("Dropping pod %q out of the queue: %v", key, err))
}

func (c *controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()

	go c.informer.Run(stopCh)

	log.Info().
		Msg(fmt.Sprintf("waiting for %s controller cache to by sync", c.resourceType))

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	log.Info().
		Msg(fmt.Sprintf("%s controller is ready and starting", c.resourceType))

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Info().
		Str("type", c.resourceType).
		Msg("Stopping controller")
}

func (c *controller) runWorker() {
	for c.processNextItem() {
	}
}

// this function should be used by all receivers in order to to send
// the updated events in parallel
//  * receiverEvent: is the new event we want to notify the receivers about
//  * receiversSlice is the slice of strings that contains the desired receiver names
func sendEventToReceivers(receiverEvent receivers.ReceiverEvent, receiversSlice []string) {
	var channelList []chan error

	for _, receiverName := range receiversSlice {
		channel := make(chan error)
		channelList = append(channelList, channel)

		if receivers.ReceiverMap[receiverName] != nil {
			go receivers.ReceiverMap[receiverName].HandleEvent(receiverEvent, channel)
		} else {
			log.Warn().Msg(fmt.Sprintf("an event was requested to be send to unknown receiver: %s", receiverName))
		}
	}

	waitForChannelsToClose(channelList...)

	// act as a default receiver. the event will
	// be logged only when running with debug log level
	reStr, _ := json.Marshal(receiverEvent)
	log.Debug().Msg(string(reStr))
}

func waitForChannelsToClose(chans ...chan error) {
	t := time.Now()
	for _, v := range chans {
		if err := <-v; err != nil {
			log.Error().Msg(fmt.Sprintf("an error occuerd during send event to a receiver: %s", err))
		} else {
			log.Debug().Msg(fmt.Sprintf("%v for chan to close\n", time.Since(t)))
		}
	}

	log.Debug().Msg(fmt.Sprintf("%v for channels to close\n", time.Since(t)))
}

func init() {
	log.Info().Msg("initializing k8s client")
	// by default, we are trying to initalize 'in cluster' client,
	// if error occuer we fallback to 'out of cluster' client
	config, err := rest.InClusterConfig()

	if err != nil {
		// out of cluster
		k8sClient.Clientset = initClientOutOfCluster()
		log.Info().Msg("k8s 'out of cluster' client is initialized")
	} else {
		// in cluster
		clientset, err := kubernetes.NewForConfig(config)

		if err != nil {
			panic(err.Error())
		}

		k8sClient.Clientset = clientset
		log.Info().Msg("k8s 'in cluster' client is initialized")
	}
}

// StartWatch function is used to trigger our watchers for k8s resources
func StartWatch(initTime time.Time) {
	applicationInitTime = initTime

	// podController := newPodController() // pod watcher
	hpaController := newHPAController() // Horizontal Pod Autoscaler watcher

	stopCh := make(chan struct{})
	defer close(stopCh)

	// run controllers
	// go podController.Run(1, stopCh)
	go hpaController.Run(1, stopCh)

	// wait forever
	select {}
}
