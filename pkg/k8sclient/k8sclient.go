package k8sclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sClient *kubernetes.Clientset

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h // linux
	}

	return os.Getenv("USERPROFILE") // windows
}

func initClientOutOfCluster() *kubernetes.Clientset {
	var kubeconfig *string

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

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

func init() {
	// by default, we are trying to initalize 'in cluster' client,
	// if error occuer we fallback to 'out of cluster' client
	config, err := rest.InClusterConfig()

	if err != nil {
		// out of cluster
		k8sClient = initClientOutOfCluster()
	} else {
		// in cluster
		clientset, err := kubernetes.NewForConfig(config)

		if err != nil {
			panic(err.Error())
		}

		k8sClient = clientset
	}

	watchlist := cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(), "pods", "logs", fields.Everything())

	_, controller := cache.NewInformer(watchlist, &v1.Pod{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("add: %s \n", obj.(*v1.Pod).ObjectMeta.Name)
				// fmt.Printf("add: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("delete: %s \n", obj.(*v1.Pod).ObjectMeta.Name)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("\n %s in namespace %s has been updated from %s to %s \n", newObj.(*v1.Pod).ObjectMeta.Name, newObj.(*v1.Pod).ObjectMeta.Namespace, oldObj.(*v1.Pod).Status.Phase, newObj.(*v1.Pod).Status.Phase)
				formatConditionsArray(newObj.(*v1.Pod))
				//fmt.Printf("\n The update was initiated since %s's state changed from %s to %s \n", newObj.(*v1.Pod).ObjectMeta.Name, oldObj.(*v1.Pod).Status.Conditions, newObj.(*v1.Pod).Status.Conditions)
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	fmt.Println("Done with init k8s client")
	// Wait forever
	select {}
}

// Client is getter function for k8sclient
func Client() *kubernetes.Clientset {
	return k8sClient
}

func formatConditionsArray(p *v1.Pod) {
	var podConditionsArray []v1.PodCondition = p.Status.Conditions
	var initContainerStatusArray []v1.ContainerStatus = p.Status.InitContainerStatuses

	for _, c := range podConditionsArray {
		if c.Type == "Initialized" && c.Status != "True" {
			var initContainersNames string = getStringInBetween(c.Message, "[", "]")
			fmt.Printf("\n %s's init-container(s) %s has moved to an incomplete status \n", p.ObjectMeta.Name, initContainersNames)
			parseContainerStatus(initContainerStatusArray)

		} else if c.Type == "Ready" && c.Status != "True" {
			var containersNames string = getStringInBetween(c.Message, "[", "]")
			fmt.Printf("\n %s's container(s) %s has moved to an incomplete status \n", p.ObjectMeta.Name, containersNames)
		}
	}
}

func parsePodCondition(pc v1.PodCondition) {
	//
}

func parseContainerStatus(cs []v1.ContainerStatus) {
	for _, s := range cs {
		if s.Ready != true {
			var containerCurrentStateString string = parseContainerState(s.State)
			var containerLastStateString string = parseContainerState(s.LastTerminationState)
			fmt.Printf("Container %s has terminated after %d restarts due to %s, currently in %s state", s.Name, s.RestartCount, containerLastStateString, containerCurrentStateString)
		}
	}
}

func parseContainerState(cs v1.ContainerState) string {
	var s string

	if cs.Waiting != nil {
		//
	} else if cs.Running != nil {

	} else if cs.Terminated != nil {

	}

	return s
}

// getStringInBetween Returns empty string if no start string found
func getStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)

	if s == -1 {
		return
	}

	s += len(start)
	e := strings.Index(str, end)

	if e == -1 {
		return
	}

	return str[s:e]
}
