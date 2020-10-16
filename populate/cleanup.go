package populate

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/matthewmcnew/kpdemo/defaults"
	"github.com/matthewmcnew/kpdemo/k8s"
)

func Cleanup() error {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}

	k8sclient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return errors.Wrapf(err, "building kubeconfig")
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}

	images, err := client.KpackV1alpha1().Images(defaults.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Removing %d images\n", len(images.Items))

	for _, i := range images.Items {
		deleteBackground := metav1.DeletePropagationBackground
		err := client.KpackV1alpha1().Images(defaults.Namespace).Delete(i.Name, &metav1.DeleteOptions{
			PropagationPolicy: &deleteBackground,
		})
		if err != nil {
			return err
		}
	}

	err = k8sclient.CoreV1().Namespaces().Delete(defaults.Namespace, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = deleteStack(client)
	if err != nil {
		return err
	}

	err = deleteStore(client)
	if err != nil {
		return err
	}

	return deleteBuilder(client)
}

func deleteStack(client *versioned.Clientset) error {
	stack, err := client.KpackV1alpha1().ClusterStacks().Get(defaults.StackName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	oldSpec, ok := stack.Annotations[defaults.OldSpecAnnotation]
	if ok {
		stackSpec := v1alpha1.ClusterStackSpec{}
		err := json.Unmarshal([]byte(oldSpec), &stackSpec)
		if err != nil {
			return err
		}
		delete(stack.Annotations, defaults.OldSpecAnnotation)
		stack.Spec = stackSpec

		_, err = client.KpackV1alpha1().ClusterStacks().Update(stack)
	} else {
		err = client.KpackV1alpha1().ClusterStacks().Delete(defaults.StackName, &metav1.DeleteOptions{})
	}
	return err
}

func deleteStore(client *versioned.Clientset) error {
	store, err := client.KpackV1alpha1().ClusterStores().Get(defaults.StoreName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	oldSpec, ok := store.Annotations[defaults.OldSpecAnnotation]
	if ok {
		storeSpec := v1alpha1.ClusterStoreSpec{}
		err := json.Unmarshal([]byte(oldSpec), &storeSpec)
		if err != nil {
			return err
		}
		delete(store.Annotations, defaults.OldSpecAnnotation)
		store.Spec = storeSpec

		_, err = client.KpackV1alpha1().ClusterStores().Update(store)
	} else {
		err = client.KpackV1alpha1().ClusterStores().Delete(defaults.StoreName, &metav1.DeleteOptions{})
	}
	return err
}

func deleteBuilder(client *versioned.Clientset) error {
	builder, err := client.KpackV1alpha1().ClusterBuilders().Get(defaults.ClusterBuilderName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	oldSpec, ok := builder.Annotations[defaults.OldSpecAnnotation]
	if ok {
		builderSpec := v1alpha1.ClusterBuilderSpec{}
		err := json.Unmarshal([]byte(oldSpec), &builderSpec)
		if err != nil {
			return err
		}
		delete(builder.Annotations, defaults.OldSpecAnnotation)
		builder.Spec = builderSpec

		_, err = client.KpackV1alpha1().ClusterBuilders().Update(builder)
	} else {
		err = client.KpackV1alpha1().ClusterBuilders().Delete(defaults.ClusterBuilderName, &metav1.DeleteOptions{})
	}
	return err
}
