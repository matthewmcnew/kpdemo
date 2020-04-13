package populate

import (
	"encoding/json"
	"fmt"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/matthewmcnew/pbdemo/defaults"
	"github.com/matthewmcnew/pbdemo/k8s"
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

	images, err := client.BuildV1alpha1().Images(defaults.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Removing %d images\n", len(images.Items))

	for _, i := range images.Items {
		deleteBackground := metav1.DeletePropagationBackground
		err := client.BuildV1alpha1().Images(defaults.Namespace).Delete(i.Name, &metav1.DeleteOptions{
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
	stack, err := client.ExperimentalV1alpha1().Stacks().Get(defaults.StackName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	oldSpec, ok := stack.Annotations[defaults.OldSpecAnnotation]
	if ok {
		stackSpec := expv1alpha1.StackSpec{}
		err := json.Unmarshal([]byte(oldSpec), &stackSpec)
		if err != nil {
			return err
		}
		delete(stack.Annotations, defaults.OldSpecAnnotation)
		stack.Spec = stackSpec

		_, err = client.ExperimentalV1alpha1().Stacks().Update(stack)
	} else {
		err = client.ExperimentalV1alpha1().Stacks().Delete(defaults.StackName, &metav1.DeleteOptions{})
	}
	return err
}

func deleteStore(client *versioned.Clientset) error {
	store, err := client.ExperimentalV1alpha1().Stores().Get(defaults.StoreName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	oldSpec, ok := store.Annotations[defaults.OldSpecAnnotation]
	if ok {
		storeSpec := expv1alpha1.StoreSpec{}
		err := json.Unmarshal([]byte(oldSpec), &storeSpec)
		if err != nil {
			return err
		}
		delete(store.Annotations, defaults.OldSpecAnnotation)
		store.Spec = storeSpec

		_, err = client.ExperimentalV1alpha1().Stores().Update(store)
	} else {
		err = client.ExperimentalV1alpha1().Stores().Delete(defaults.StoreName, &metav1.DeleteOptions{})
	}
	return err
}

func deleteBuilder(client *versioned.Clientset) error {
	builder, err := client.ExperimentalV1alpha1().CustomClusterBuilders().Get(defaults.BuilderName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	oldSpec, ok := builder.Annotations[defaults.OldSpecAnnotation]
	if ok {
		builderSpec := expv1alpha1.CustomClusterBuilderSpec{}
		err := json.Unmarshal([]byte(oldSpec), &builderSpec)
		if err != nil {
			return err
		}
		delete(builder.Annotations, defaults.OldSpecAnnotation)
		builder.Spec = builderSpec

		_, err = client.ExperimentalV1alpha1().CustomClusterBuilders().Update(builder)
	} else {
		err = client.ExperimentalV1alpha1().CustomClusterBuilders().Delete(defaults.BuilderName, &metav1.DeleteOptions{})
	}
	return err
}
