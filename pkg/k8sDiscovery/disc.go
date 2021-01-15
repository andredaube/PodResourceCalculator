package k8sDiscovery

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func K8s(kubeconfig string) (kubernetes.Interface, *rest.Config, error) {
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster == true {
		log.Infof("inside cluster, using in-cluster configuration")
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Errorf("Failed to get incluster config:%v", err)
			return nil, nil, err
		}
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Errorf("Failed to construct the clientSet:%v", err)
			return nil, nil, err
		}
		return clientSet, config, nil
	}

	log.Infof("outside of cluster")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Errorf("Failed to build the config:%v", err)
		return nil, nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("Failed to construct the clientSet:%v", err)
		return nil, nil, err
	}
	return clientSet, config, nil
}

//for testing
func GetServerVersion(clientSet kubernetes.Interface) (string, error) {
	version, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		log.Errorf("Failed to get server version:%v", err)
		return "", err
	}
	return fmt.Sprintf("%s", version), nil
}
