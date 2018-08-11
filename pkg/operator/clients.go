package flokkroperator

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	owncli "github.com/flokkr/flokkr-operator/pkg/clientset/versioned/typed/flokkr/v1alpha1"
	"path/filepath"
	"flag"
	"os"
)

func createKubernetesClients() (*owncli.FlokkrV1alpha1Client, *apiextensionscli.Clientset, *kubernetes.Clientset, error) {
	var err error
	var cfg *rest.Config

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if _, err := os.Stat(*kubeconfig); os.IsNotExist(err) {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}

	if err != nil {
		panic(err.Error())
	}

	// Create clients.
	k8sCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	// App CRD k8s types client.
	ownCli, err := owncli.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	// CRD cli.
	aexCli, err := apiextensionscli.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	return ownCli, aexCli, k8sCli, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
