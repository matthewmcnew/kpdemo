package buildpacks

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/matthewmcnew/build-service-visualization/defaults"
	"github.com/matthewmcnew/build-service-visualization/k8s"
)

func UpdateBuildpack(id string) error {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		return err
	}

	builder, err := client.ExperimentalV1alpha1().CustomClusterBuilders().Get(defaults.BuilderName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if !isInBuilder(builder, id) {
		return fmt.Errorf("%s is not in builder order definition use a buildpack in order:\n\n%s", id, prettyPrint(builder.Spec.Order))
	}

	store, err := client.ExperimentalV1alpha1().Stores().Get(defaults.StoreName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	storeBuildpack, err := findBuildpack(store.Status.Buildpacks, id)
	if err != nil {
		return err
	}

	reference, err := name.ParseReference(builder.Spec.Tag)
	if err != nil {
		return err
	}

	newVersion, err := newVersion(storeBuildpack.Id, storeBuildpack.Version)
	if err != nil {
		return err
	}

	fmt.Printf("Creating new buildpack %s with version %s \n", id, newVersion)
	fmt.Printf("\n This will take a moment...\n")

	newBp, err := buildpackage(id, storeBuildpack.Version, storeBuildpack.StoreImage.Image, fmt.Sprintf("%s/%s:%s", reference.Context().RegistryStr(), reference.Context().RepositoryStr(), newVersion))
	if err != nil {
		return err
	}

	fmt.Printf("wrote to %s \n", newBp)

	store.Spec.Sources = append(store.Spec.Sources, v1alpha1.StoreImage{Image: newBp})
	_, err = client.ExperimentalV1alpha1().Stores().Update(store)
	return err
}

func isInBuilder(builder *v1alpha1.CustomClusterBuilder, id string) bool {
	for _, oe := range builder.Spec.Order {
		for _, bp := range oe.Group {
			if bp.Id == id {
				return true
			}
		}
	}
	return false
}

func findBuildpack(storeBps []v1alpha1.StoreBuildpack, id string) (v1alpha1.StoreBuildpack, error) {
	var matchingBuildpacks []v1alpha1.StoreBuildpack
	for _, buildpack := range storeBps {
		if buildpack.Id == id {
			matchingBuildpacks = append(matchingBuildpacks, buildpack)
		}
	}

	if len(matchingBuildpacks) == 0 {
		return v1alpha1.StoreBuildpack{}, errors.Errorf("could not find buildpack with id '%s'", id)
	}

	return highestVersion(matchingBuildpacks)
}

func highestVersion(matchingBuildpacks []v1alpha1.StoreBuildpack) (v1alpha1.StoreBuildpack, error) {
	for _, bp := range matchingBuildpacks {
		if _, err := semver.NewVersion(bp.Version); err != nil {
			return v1alpha1.StoreBuildpack{}, errors.Errorf("cannot find buildpack '%s' with latest version due to invalid semver '%s'", bp.Id, bp.Version)
		}
	}
	sort.Sort(byBuildpackVersion(matchingBuildpacks))
	return matchingBuildpacks[len(matchingBuildpacks)-1], nil
}

type byBuildpackVersion []v1alpha1.StoreBuildpack

func (b byBuildpackVersion) Len() int {
	return len(b)
}

func (b byBuildpackVersion) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byBuildpackVersion) Less(i, j int) bool {
	return semver.MustParse(b[i].Version).LessThan(semver.MustParse(b[j].Version))
}

func prettyPrint(order []v1alpha1.OrderEntry) string {
	sb := strings.Builder{}
	for _, oe := range order {
		sb.WriteString("Order:\n")
		for _, bp := range oe.Group {
			sb.WriteString(fmt.Sprintf("  %s\n", bp.Id))
		}
	}
	return sb.String()
}
