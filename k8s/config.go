package k8s

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func BuildConfigFromFlags(masterURL, kubeconfigPath string) (*rest.Config, error) {

	var clientConfigLoader clientcmd.ClientConfigLoader

	if kubeconfigPath == "" {
		clientConfigLoader = clientcmd.NewDefaultClientConfigLoadingRules()
	} else {
		clientConfigLoader = &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientConfigLoader,
		&clientcmd.ConfigOverrides{ClusterInfo: api.Cluster{Server: masterURL}}).ClientConfig()

}
