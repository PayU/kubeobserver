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
				fmt.Printf("old: %s, new: %s \n", oldObj.(*v1.Pod).ObjectMeta.Name, newObj.(*v1.Pod).ObjectMeta.Name)
				var oldInitContainersNames []string

				for _, c := range oldObj.(*v1.Pod).Spec.InitContainers {
					oldInitContainersNames := append(oldInitContainersNames, c.Name)
					fmt.Printf("This is old %s's init containers: %v \n", oldObj.(*v1.Pod).ObjectMeta.Name, oldInitContainersNames)
				}

				//fmt.Printf("init containers: %v", oldObj.(*v1.Pod).Spec.InitContainers)
				//fmt.Printf("containers: %v", oldObj.(*v1.Pod).Spec.Containers)
				//fmt.Println("old container status: ", "new container status: ", oldObj.(*v1.Pod).Status.ContainerStatuses, newObj.(*v1.Pod).Status.ContainerStatuses)
				//fmt.Println("old init container status: ", "new init container status: ", oldObj.(*v1.Pod).Status.InitContainerStatuses, newObj.(*v1.Pod).Status.InitContainerStatuses)
				//fmt.Printf("old state: %s, new state: %s \n", oldObj.(*v1.Pod).Status.Phase, newObj.(*v1.Pod).Status.Phase)
				//fmt.Printf("old condition: %s, new condition: %s \n", oldObj.(*v1.Pod).Status.Conditions, newObj.(*v1.Pod).Status.Conditions)
				fmt.Println("")
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
