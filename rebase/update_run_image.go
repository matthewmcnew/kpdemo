package rebase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
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

	builder, err := client.BuildV1alpha1().ClusterBuilders().Get(defaults.BuilderName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if builder.Status.Stack.RunImage == "" {
		return errors.New("error parsing builder run image")
	}

	reference, err := name.ParseReference(builder.Status.Stack.RunImage)
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

	i, err = imagehelpers.SetStringLabel(i, "BUILD_SERVICE_DEMO", time.Now().String())
	if err != nil {
		return err
	}

	err = remote.Write(updateRef, i, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return err
	}

	digest, err := i.Digest()
	if err != nil {
		return err
	}
	fmt.Printf("Updated Run Image %s@%s\n", updateRef, digest)

	builder.Annotations = map[string]string{
		"BUILD_SERVICE_DEMO": time.Now().String(),
	}

	_, err = client.BuildV1alpha1().ClusterBuilders().Update(builder)

	return err
}
