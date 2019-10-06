package rebase

import (
	"errors"
	"fmt"
	"github.com/buildpack/imgutil/remote"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/matthewmcnew/build-service-visualization/defaults"
	"github.com/matthewmcnew/build-service-visualization/k8s"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
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

	if builder.Status.RunImage == "" {
		return errors.New("error parsing builder run image")
	}

	reference, err := name.ParseReference(builder.Status.RunImage)
	if err != nil {
		return err
	}

	runImage := fmt.Sprintf("%s/%s:pbdemo", reference.Context().RegistryStr(), reference.Context().RepositoryStr())
	fmt.Printf("Pushing update to: %s\n", runImage)

	image, err := remote.NewImage(runImage, authn.DefaultKeychain, remote.FromBaseImage(builder.Status.RunImage))
	if err != nil {
		return err
	}

	err = image.SetLabel("BUILD_SERVICE_DEMO", time.Now().String())
	if err != nil {
		return err
	}

	err = image.Save()
	if err != nil {
		return err
	}

	identifier, err := image.Identifier()
	if err != nil {
		return err
	}
	fmt.Printf("Updated Run Image %s\n", identifier)

	builder.Annotations = map[string]string{
		"BUILD_SERVICE_DEMO": time.Now().String(),
	}

	_, err = client.BuildV1alpha1().ClusterBuilders().Update(builder)

	return err
}
