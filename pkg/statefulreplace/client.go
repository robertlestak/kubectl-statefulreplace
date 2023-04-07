package statefulreplace

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	K8sClient *kubernetes.Clientset
)

// CreateKubeClient creates a global k8s client
func CreateKubeClient() error {
	l := log.WithFields(
		log.Fields{
			"action": "CreateKubeClient",
		},
	)
	l.Debug("get CreateKubeClient")
	var kubeconfig string
	var err error
	if os.Getenv("KUBECONFIG") != "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	var config *rest.Config
	// naïvely assume if no kubeconfig file that we are running in cluster
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		config, err = rest.InClusterConfig()
		if err != nil {
			l.Errorf("res.InClusterConfig error=%v", err)
			return err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			l.Errorf("clientcmd.BuildConfigFromFlags error=%v", err)
			return err
		}
	}
	K8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		l.Errorf("kubernetes.NewForConfig error=%v", err)
		return err
	}
	return nil
}

func init() {
	if err := CreateKubeClient(); err != nil {
		log.Fatalf("CreateKubeClient error=%v", err)
	}
}
