package k8sclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var k8sClient *kubernetes.Clientset
var prevNewPod *v1.Pod

type podStatusByConditions struct {
	Init        bool
	Ready       bool
	ContainersR bool
	Sched       bool
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
				fmt.Println("start of update func")
				//equal := podByConditionsEqual(oldObj.(*v1.Pod), newObj.(*v1.Pod))
				//oldLowTime, oldHighTime := sortPodTimeStamps(oldObj.(*v1.Pod))
				//newLowTime, newHighTime := sortPodTimeStamps(newObj.(*v1.Pod))

				//fmt.Printf("old low: %v, old high: %v, new low: %v, new high: %v \n", oldLowTime, oldHighTime, newLowTime, newHighTime)

				//if !equal || prevNewPod == nil {
				fmt.Printf("Pod %s in namespace %s has changed at %v\n", oldObj.(*v1.Pod).ObjectMeta.Name, oldObj.(*v1.Pod).ObjectMeta.Namespace, time.Now())
				//formatConditionsArray(oldObj.(*v1.Pod))
				formatConditionsArray(newObj.(*v1.Pod))
				fmt.Println("end of update func")

				//}

				prevNewPod = newObj.(*v1.Pod)

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
	pod := representPodByConditions(p)

	if pod.Sched && !pod.Init && !pod.ContainersR && !pod.Ready {
		fmt.Printf("Pod %s's init-containers changed at %v: \n", p.ObjectMeta.Name, time.Now())
		parseContainerStatus(p.Status.InitContainerStatuses, p.Status.Message, p.Status.Reason)
	} else if pod.Sched && pod.Init && !pod.ContainersR && !pod.Ready {
		fmt.Printf("Pod %s's containers changed at %v: \n", p.ObjectMeta.Name, time.Now())
		parseContainerStatus(p.Status.ContainerStatuses, p.Status.Message, p.Status.Reason)
	}

}

func parseContainerStatus(cs []v1.ContainerStatus, message string, reason string) {
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

func podByConditionsEqual(op *v1.Pod, np *v1.Pod) bool {
	oldPod := representPodByConditions(op)
	newPod := representPodByConditions(np)

	return cmp.Equal(oldPod, newPod)
}

func representPodByConditions(p *v1.Pod) podStatusByConditions {
	var podConditionsArray []v1.PodCondition = p.Status.Conditions
	pod := podStatusByConditions{false, false, false, false}

	for _, c := range podConditionsArray {
		if c.Type == "Initialized" && c.Status == "True" {
			fmt.Printf("message: %s reason: %s \n", c.Message, c.Reason)
			pod.Init = true
		} else if c.Type == "Ready" && c.Status == "True" {
			fmt.Printf("message: %s reason: %s \n", c.Message, c.Reason)
			pod.Ready = true
		} else if c.Type == "ContainersReady" && c.Status == "True" {
			fmt.Printf("message: %s reason: %s \n", c.Message, c.Reason)
			pod.ContainersR = true
		} else if c.Type == "PodScheduled" && c.Status == "True" {
			fmt.Printf("message: %s reason: %s \n", c.Message, c.Reason)
			pod.Sched = true
		}
	}

	return pod
}

func sortPodTimeStamps(p *v1.Pod) (metav1.Time, metav1.Time) {
	var podConditionsArray []v1.PodCondition = p.Status.Conditions
	//fmt.Printf("podConditions are: %v \n", podConditionsArray)
	var highTime metav1.Time
	var lowTime metav1.Time

	if len(podConditionsArray) == 0 {
		highTime = metav1.Now()
		lowTime = metav1.Now()
		//fmt.Printf("no relevant times: low time: %v \t high time: %v \n", lowTime, highTime)
	} else if len(podConditionsArray) > 0 {
		highTime = podConditionsArray[0].LastTransitionTime
		lowTime = podConditionsArray[0].LastTransitionTime

		for _, c := range podConditionsArray {
			if c.LastTransitionTime.Before(&lowTime) && c.Type != "PodScheduled" {
				lowTime = c.LastTransitionTime
			} else if highTime.Before(&c.LastTransitionTime) && c.Type != "PodScheduled" {
				highTime = c.LastTransitionTime
			}
		}
	}

	//fmt.Printf("low time: %v \t high time: %v \n", lowTime, highTime)
	return lowTime, highTime
}
