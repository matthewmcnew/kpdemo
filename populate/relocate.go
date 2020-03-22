package populate

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/matthewmcnew/build-service-visualization/defaults"
	"github.com/matthewmcnew/build-service-visualization/k8s"
)

type Relocated struct {
	Order expv1alpha1.Order
}

func Relocate(imageTag string) (Relocated, error) {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return Relocated{}, errors.Wrapf(err, "building kubeconfig")
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		return Relocated{}, errors.Wrapf(err, "building kubeconfig")
	}

	runRef, err := name.ParseReference("cloudfoundry/run:base-cnb")
	if err != nil {
		return Relocated{}, err
	}

	run, err := remote.Image(runRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return Relocated{}, err
	}

	relocatedRunRef, err := name.ParseReference(imageTag + ":run")
	if err != nil {
		return Relocated{}, err
	}

	runImage, err := save(relocatedRunRef, run)
	if err != nil {
		return Relocated{}, err
	}

	builderRef, err := name.ParseReference("cloudfoundry/cnb:bionic")
	if err != nil {
		return Relocated{}, err
	}

	builder, err := remote.Image(builderRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return Relocated{}, err
	}

	var order []expv1alpha1.OrderEntry
	err = imagehelpers.GetLabel(builder, "io.buildpacks.buildpack.order", &order)
	if err != nil {
		return Relocated{}, err
	}

	relocatedBuilderRef, err := name.ParseReference(imageTag + ":buildpacks")
	if err != nil {
		return Relocated{}, err
	}
	buildpacksImage, err := save(relocatedBuilderRef, builder)

	err = saveStack(client, &expv1alpha1.Stack{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaults.StackName,
		},
		Spec: expv1alpha1.StackSpec{
			Id: "io.buildpacks.stacks.bionic",
			BuildImage: expv1alpha1.StackSpecImage{
				Image: "cloudfoundry/build:base-cnb",
			},
			RunImage: expv1alpha1.StackSpecImage{
				Image: runImage,
			},
		},
	})
	if err != nil {
		return Relocated{}, err
	}

	err = saveStore(client, &expv1alpha1.Store{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaults.StoreName,
		},
		Spec: expv1alpha1.StoreSpec{
			Sources: []expv1alpha1.StoreImage{
				{
					Image: buildpacksImage,
				},
			},
		},
	})

	return Relocated{
		Order: order,
	}, err
}

func save(ref name.Reference, i v1.Image) (string, error) {
	err := remote.Write(ref, i, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", err
	}

	digest, err := i.Digest()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s@%s", ref.Name(), digest.String()), nil
}

func saveStore(client *versioned.Clientset, store *expv1alpha1.Store) error {
	existingStore, err := client.ExperimentalV1alpha1().Stores().Get(defaults.StoreName, metav1.GetOptions{})
	if err != nil && !k8errors.IsNotFound(err) {
		return err
	}
	if k8errors.IsNotFound(err) {
		_, err = client.ExperimentalV1alpha1().Stores().Create(store)
	} else {
		oldSpec, err := json.Marshal(existingStore.Spec)
		if err != nil {
			return err
		}

		if existingStore.Annotations == nil {
			existingStore.Annotations = map[string]string{}
		}

		existingStore.Annotations[defaults.OldSpecAnnotation] = string(oldSpec)
		existingStore.Spec = store.Spec
		_, err = client.ExperimentalV1alpha1().Stores().Update(existingStore)
	}
	return err
}

func saveStack(client *versioned.Clientset, stack *expv1alpha1.Stack) error {
	existingStack, err := client.ExperimentalV1alpha1().Stacks().Get(defaults.StackName, metav1.GetOptions{})
	if err != nil && !k8errors.IsNotFound(err) {
		return err
	}
	if k8errors.IsNotFound(err) {
		_, err = client.ExperimentalV1alpha1().Stacks().Create(stack)
	} else {
		oldSpec, err := json.Marshal(existingStack.Spec)
		if err != nil {
			return err
		}

		if existingStack.Annotations == nil {
			existingStack.Annotations = map[string]string{}
		}

		existingStack.Annotations[defaults.OldSpecAnnotation] = string(oldSpec)
		existingStack.Spec = stack.Spec
		_, err = client.ExperimentalV1alpha1().Stacks().Update(existingStack)
	}
	return err
}
