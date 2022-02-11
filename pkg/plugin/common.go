package plugin

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type NamespaceResult struct {
	Value v1.Namespace
	Err   error
}

type PodResult struct {
	Value v1.Pod
	Err   error
}

// GetClientset: get kubernetes clientset for interacting with the kubernetes cluster
func GetClientset(configFlags *genericclioptions.ConfigFlags) (*kubernetes.Clientset, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, nil
}
