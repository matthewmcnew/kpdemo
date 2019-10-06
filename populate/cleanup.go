package populate

import (
	"fmt"
	"github.com/matthewmcnew/build-service-visualization/defaults"
	"github.com/matthewmcnew/build-service-visualization/k8s"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Cleanup() error {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}

	images, err := client.BuildV1alpha1().Images(defaults.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Removing %d images\n", len(images.Items))

	for _, i := range images.Items {
		err := client.BuildV1alpha1().Images(defaults.Namespace).Delete(i.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return client.BuildV1alpha1().ClusterBuilders().Delete(defaults.BuilderName, &metav1.DeleteOptions{})
}
