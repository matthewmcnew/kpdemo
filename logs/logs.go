package logs

import (
	"context"
	"os"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/logs"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/matthewmcnew/pbdemo/k8s"
)

func Logs(name string) error {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		return errors.Wrapf(err, "building kubeconfig")
	}

	list, err := client.BuildV1alpha1().Images("").List(v1.ListOptions{})
	if err != nil {
		return err
	}

	image, ok := find(list, name)
	if !ok {
		return errors.Errorf("could not find image: %s", name)
	}

	return logs.NewBuildLogsClient(k8sClient).Tail(context.Background(), os.Stdout, image.Name, "", image.Namespace)
}

func find(list *v1alpha1.ImageList, name string) (*v1alpha1.Image, bool) {
	for _, i := range list.Items {
		if i.Name == name {
			return &i, true
		}
	}
	return nil, false
}
