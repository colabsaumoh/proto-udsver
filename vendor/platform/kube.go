package platform

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	restclient "k8s.io/client-go/rest"

)

type K8sQuery struct {
	ContainerId	string
	RespChan	chan<- string
}

const (
	staticNodeName string = "test-istio"
)

func getPodInfo(cset *kubernetes.Clientset, cid *string, nodeName string) (string) {
	qStr := "spec.nodeName=" + nodeName
	opts := metav1.ListOptions{}
	opts.FieldSelector = qStr
	pods, err := cset.CoreV1().Pods("").List(opts)
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("Number of pods on %v %v", qStr, len(pods.Items))

	matchCid := "docker://" + *cid
	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.ContainerID == matchCid {
				return pod.ObjectMeta.Name
			}
		}
	}
	return ""
}

// Invoked as a go routine
func QueryServer(inCluster bool, kubeconfig *string, qChan <-chan K8sQuery) {
	var config *restclient.Config
	var err error

	if inCluster == false {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Fatalf("%v, Failed to get the k8s config %v", kubeconfig, err.Error())
		}
	} else {
		config, err = restclient.InClusterConfig()
		if err != nil {
			log.Fatalf("%v, Failed to get the k8s config (in cluster) %v", kubeconfig, err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("%v", err.Error())
	}

        for {
		query, more := <-qChan
		if !more {
			log.Println("no more questions on query channel.")
			return
		}
		resp := getPodInfo(clientset, &query.ContainerId, staticNodeName)
		query.RespChan<- resp
		close(query.RespChan)
	}
}
