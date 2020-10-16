package populate

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/matthewmcnew/kpdemo/defaults"
	"github.com/matthewmcnew/kpdemo/k8s"
)

const (
	kpackNamespace = "kpack"
	kpConfigMap    = "kp-config"
	canonRepoKey   = "canonical.repository"
)

type Relocated struct {
	Order v1alpha1.Order
}

func Relocate(imageTag string) (Relocated, error) {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return Relocated{}, errors.Wrapf(err, "building kubeconfig")
	}

	k8sClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return Relocated{}, errors.Wrapf(err, "building kubeconfig")
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		return Relocated{}, errors.Wrapf(err, "building kubeconfig")
	}

	runRef, err := name.ParseReference("paketobuildpacks/run:base-cnb")
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

	if !verifyRegistryPublic(k8sClient, runImage) {
		fmt.Printf("\n%s: Image: %s is not public. \n kpdemo populate will not work if %s is not public or readable by kpack and the nodes on the cluster\n Continuing anyway...\n\n",
			color.RedString("WARNING"),
			imageTag,
			imageTag)
	}

	builderRef, err := name.ParseReference("paketobuildpacks/builder:base")
	if err != nil {
		return Relocated{}, err
	}

	builder, err := remote.Image(builderRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return Relocated{}, err
	}

	var order []v1alpha1.OrderEntry
	err = imagehelpers.GetLabel(builder, "io.buildpacks.buildpack.order", &order)
	if err != nil {
		return Relocated{}, err
	}

	relocatedBuilderRef, err := name.ParseReference(imageTag + ":buildpacks")
	if err != nil {
		return Relocated{}, err
	}
	buildpacksImage, err := save(relocatedBuilderRef, builder)

	err = saveClusterStack(client, &v1alpha1.ClusterStack{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaults.StackName,
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: "io.buildpacks.stacks.bionic",
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: "paketobuildpacks/build:base-cnb",
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: runImage,
			},
		},
	})
	if err != nil {
		return Relocated{}, err
	}

	err = saveClusterStore(client, &v1alpha1.ClusterStore{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaults.StoreName,
		},
		Spec: v1alpha1.ClusterStoreSpec{
			Sources: []v1alpha1.StoreImage{
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

func saveClusterStore(client *versioned.Clientset, store *v1alpha1.ClusterStore) error {
	existingClusterStore, err := client.KpackV1alpha1().ClusterStores().Get(defaults.StoreName, metav1.GetOptions{})
	if err != nil && !k8errors.IsNotFound(err) {
		return err
	}
	if k8errors.IsNotFound(err) {
		_, err = client.KpackV1alpha1().ClusterStores().Create(store)
	} else {
		oldSpec, err := json.Marshal(existingClusterStore.Spec)
		if err != nil {
			return err
		}

		if existingClusterStore.Annotations == nil {
			existingClusterStore.Annotations = map[string]string{}
		}

		existingClusterStore.Annotations[defaults.OldSpecAnnotation] = string(oldSpec)
		existingClusterStore.Spec = store.Spec
		_, err = client.KpackV1alpha1().ClusterStores().Update(existingClusterStore)
	}
	return err
}

func saveClusterStack(client *versioned.Clientset, stack *v1alpha1.ClusterStack) error {
	existingClusterStack, err := client.KpackV1alpha1().ClusterStacks().Get(defaults.StackName, metav1.GetOptions{})
	if err != nil && !k8errors.IsNotFound(err) {
		return err
	}
	if k8errors.IsNotFound(err) {
		_, err = client.KpackV1alpha1().ClusterStacks().Create(stack)
	} else {
		oldSpec, err := json.Marshal(existingClusterStack.Spec)
		if err != nil {
			return err
		}

		if existingClusterStack.Annotations == nil {
			existingClusterStack.Annotations = map[string]string{}
		}

		existingClusterStack.Annotations[defaults.OldSpecAnnotation] = string(oldSpec)
		existingClusterStack.Spec = stack.Spec
		_, err = client.KpackV1alpha1().ClusterStacks().Update(existingClusterStack)
	}
	return err
}

func verifyRegistryPublic(client *kubernetes.Clientset, image string) bool {
	if isBuildServiceRegistry(client, image) {
		return true
	}

	ref, _ := name.ParseReference(image)

	_, err := remote.Image(ref, remote.WithAuth(authn.Anonymous))
	if err != nil {
		return false
	}

	return true
}

func isBuildServiceRegistry(client *kubernetes.Clientset, image string) bool {
	kpConfig, err := client.CoreV1().ConfigMaps(kpackNamespace).Get(kpConfigMap, metav1.GetOptions{})
	if err != nil {
		return false
	}

	canonRepo, ok := kpConfig.Data[canonRepoKey]
	if !ok {
		return false
	}

	canonReg, err := name.ParseReference(canonRepo)
	if err != nil {
		return false
	}

	reg, err := name.ParseReference(image)
	if err != nil {
		return false
	}

	return canonReg.Context().RegistryStr() == reg.Context().RegistryStr()
}
