package rebase

import (
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/matthewmcnew/build-service-visualization/defaults"
	"github.com/matthewmcnew/build-service-visualization/k8s"
)

func UpdateRunImage() error {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}

	stack, err := client.ExperimentalV1alpha1().Stacks().Get(defaults.StackName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	reference, err := name.ParseReference(stack.Spec.RunImage.Image)
	if err != nil {
		return err
	}

	updateRef, err := name.ParseReference(fmt.Sprintf("%s/%s:run", reference.Context().RegistryStr(), reference.Context().RepositoryStr()))
	if err != nil {
		return err
	}

	fmt.Printf("Pushing update to: %s\n", updateRef.Name())

	i, err := remote.Image(reference, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return err
	}

	i, err = imagehelpers.SetStringLabel(i, "PBDEMO_DEMO", time.Now().String())
	if err != nil {
		return err
	}

	updatedImage, err := save(updateRef, i)
	if err != nil {
		return err
	}

	fmt.Printf("Updated Run Image %s@%s\n", updatedImage)

	stack.Spec.RunImage.Image = updatedImage
	_, err = client.ExperimentalV1alpha1().Stacks().Update(stack)
	return err
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
