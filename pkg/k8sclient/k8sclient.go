package k8sclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sClient *kubernetes.Clientset

type podStatusByConditions struct {
	Ready       bool
	Init        bool
	Sched       bool
	ContainersR bool
	AllT        bool
}

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
	var InitTime = time.Now()

	_, controller := cache.NewInformer(watchlist, &v1.Pod{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if (InitTime).Before(obj.(*v1.Pod).ObjectMeta.CreationTimestamp.Time) {
					fmt.Printf("Pod %s has been added \n", obj.(*v1.Pod).ObjectMeta.Name)
				}
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("Pod %s has been deleted \n", obj.(*v1.Pod).ObjectMeta.Name)
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				equal, oldAllT, newAllT := compareConditionsArray(oldObj.(*v1.Pod), newObj.(*v1.Pod))

				if !equal {
					fmt.Printf("Pod %s in namespace %s has changed at %v\n", oldObj.(*v1.Pod).ObjectMeta.Name, oldObj.(*v1.Pod).ObjectMeta.Namespace, time.Now())

					if oldAllT {
						formatConditionsArray(newObj.(*v1.Pod))
					} else if newAllT {
						formatConditionsArray(oldObj.(*v1.Pod))
					} else {
						formatConditionsArray(oldObj.(*v1.Pod))
						formatConditionsArray(newObj.(*v1.Pod))
					}
				}

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
	pod := podStatusByConditions{true, true, true, true, true}

	for _, c := range podConditionsArray {
		if c.Type == "Initialized" && c.Status != "True" {
			pod.Init = false
			pod.AllT = false
		} else if c.Type == "ContainersReady" && c.Status != "True" {
			pod.ContainersR = false
			pod.AllT = false
		} else if c.Type == "PodScheduled" && c.Status != "True" {
			pod.Sched = false
			pod.AllT = false
		} else if c.Type == "Ready" && c.Status != "True" {
			pod.Ready = false
			pod.AllT = false
		}
	}

	if pod.Sched && !pod.Init && !pod.ContainersR && !pod.Ready {
		fmt.Printf("Pod %s's init-containers changed at %v: \n", p.ObjectMeta.Name, time.Now())
		parseContainerStatus(p.Status.InitContainerStatuses)
	} else if pod.Sched && pod.Init && !pod.ContainersR && !pod.Ready {
		fmt.Printf("Pod %s's containers changed at %v: \n", p.ObjectMeta.Name, time.Now())
		parseContainerStatus(p.Status.ContainerStatuses)
	}
}

func parseContainerStatus(cs []v1.ContainerStatus) {
	for _, s := range cs {
		if !s.Ready {
			var containerStateString string = parseContainerState(s.State)
			fmt.Printf("### %s. The Container had %d restarts and %s", s.Name, s.RestartCount, containerStateString)

			if s.LastTerminationState.Waiting != nil && s.LastTerminationState.Running != nil && s.LastTerminationState.Terminated != nil {
				parseContainerState(s.LastTerminationState)
				fmt.Printf("### %s. Last time this container %s it was after %d restarts", s.Name, containerStateString, s.RestartCount)
			}
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

func compareConditionsArray(op *v1.Pod, np *v1.Pod) (bool, bool, bool) {
	var oldPodConditionsArray []v1.PodCondition = op.Status.Conditions
	var newPodConditionsArray []v1.PodCondition = np.Status.Conditions
	oldPod := podStatusByConditions{true, true, true, true, true}
	newPod := podStatusByConditions{true, true, true, true, true}

	for _, c := range oldPodConditionsArray {
		if c.Type == "Initialized" && c.Status != "True" {
			oldPod.Init = false
			oldPod.AllT = false
		} else if c.Type == "ContainersReady" && c.Status != "True" {
			oldPod.ContainersR = false
			oldPod.AllT = false
		} else if c.Type == "PodScheduled" && c.Status != "True" {
			oldPod.Sched = false
			oldPod.AllT = false
		} else if c.Type == "Ready" && c.Status != "True" {
			oldPod.Ready = false
			oldPod.AllT = false
		}
	}

	for _, c := range newPodConditionsArray {
		if c.Type == "Initialized" && c.Status != "True" {
			newPod.Init = false
			newPod.AllT = false
		} else if c.Type == "ContainersReady" && c.Status != "True" {
			newPod.ContainersR = false
			newPod.AllT = false
		} else if c.Type == "PodScheduled" && c.Status != "True" {
			newPod.Sched = false
			newPod.AllT = false
		} else if c.Type == "Ready" && c.Status != "True" {
			newPod.Ready = false
			newPod.AllT = false
		}
	}

	if oldPod != newPod && oldPod.AllT && newPod.AllT {
		return false, true, true
	} else if oldPod != newPod && !oldPod.AllT && !newPod.AllT {
		return false, false, false
	}

	return true, true, true

}
