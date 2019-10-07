package logs

import (
	"context"
	"github.com/matthewmcnew/build-service-visualization/defaults"
	"github.com/matthewmcnew/build-service-visualization/k8s"
	"github.com/pivotal/kpack/pkg/logs"
	"k8s.io/client-go/kubernetes"
	"os"
)

func Logs(image string) error {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}

	return logs.NewBuildLogsClient(k8sClient).Tail(context.Background(), os.Stdout, image, "", defaults.Namespace)
}
