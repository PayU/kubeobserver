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
				if oldObj.(*v1.Pod).Status.Phase != newObj.(*v1.Pod).Status.Phase {
					fmt.Printf("%s in namespace %s has been updated from %s to %s \n", newObj.(*v1.Pod).ObjectMeta.Name, newObj.(*v1.Pod).ObjectMeta.Namespace, oldObj.(*v1.Pod).Status.Phase, newObj.(*v1.Pod).Status.Phase)
				} else if oldObj.(*v1.Pod).Status.Phase == "" {
					fmt.Printf("%s in namespace %s has been updated to %s \n", newObj.(*v1.Pod).ObjectMeta.Name, newObj.(*v1.Pod).ObjectMeta.Namespace, newObj.(*v1.Pod).Status.Phase)
				}

				//formatConditionsArray(oldObj.(*v1.Pod))
				formatConditionsArray(newObj.(*v1.Pod))
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
	var containersStatusArray []v1.ContainerStatus
	var containerFunction string

	for _, c := range podConditionsArray {
		if c.Type == "Initialized" && c.Status != "True" {
			containerFunction = "init-"
			containersStatusArray = p.Status.InitContainerStatuses
			fmt.Printf("Change in one or more of %s's %scontainers: \n", p.ObjectMeta.Name, containerFunction)
		} else if c.Type == "ContainersReady" && c.Status != "True" {
			containersStatusArray = p.Status.ContainerStatuses
			fmt.Printf("Change in one or more of %s's %scontainers: \n", p.ObjectMeta.Name, containerFunction)
		}

		if len(containersStatusArray) > 0 {
			parseContainerStatus(containersStatusArray)
		}
	}
}

func parseContainerStatus(cs []v1.ContainerStatus) {
	for _, s := range cs {
		var containerStateString string = parseContainerState(s.State)
		fmt.Printf("%s. The Container had %d restarts and %s", s.Name, s.RestartCount, containerStateString)

		if s.LastTerminationState.Waiting != nil && s.LastTerminationState.Running != nil && s.LastTerminationState.Terminated != nil {
			parseContainerState(s.LastTerminationState)
			fmt.Printf(" %s. Last time this container %s it was after %d restarts", s.Name, containerStateString, s.RestartCount)
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
