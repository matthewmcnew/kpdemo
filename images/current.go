package images

import (
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	v1alpha1Listers "github.com/pivotal/kpack/pkg/client/listers/build/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"time"
)

type Image struct {
	Name          string                         `json:"name"`
	Namespace     string                         `json:"namespace"`
	BuildCount    int64                          `json:"buildCount"`
	Status        string                         `json:"status"`
	BuildMetadata v1alpha1.BuildpackMetadataList `json:"buildMetadata"`
	Remaining     int                            `json:"remaining"`
	CreatedAt     time.Time                      `json:"createdAt"`
}

func Current(lister v1alpha1Listers.ImageLister, buildLister v1alpha1Listers.BuildLister) ([]Image, error) {
	kpackImages, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	images := make([]Image, 0, len(kpackImages))
	for _, i := range kpackImages {

		images = append(images, Image{
			Name:          i.Name,
			Namespace:     i.Namespace,
			BuildCount:    i.Status.BuildCounter,
			Status:        status(i),
			CreatedAt:     i.CreationTimestamp.Time,
			BuildMetadata: builddpacks(buildLister, i),
			Remaining:     remaining(buildLister, i),
		})
	}

	return images, nil
}

func builddpacks(buildLister v1alpha1Listers.BuildLister, image *v1alpha1.Image) v1alpha1.BuildpackMetadataList {
	if image.Status.LatestBuildRef == "" {
		return nil
	}

	build, err := buildLister.Builds(image.Namespace).Get(image.Status.LatestBuildRef)
	if err != nil {
		return nil
	}

	return build.Status.BuildMetadata
}

func remaining(buildLister v1alpha1Listers.BuildLister, image *v1alpha1.Image) int {
	if image.Status.LatestBuildRef == "" {
		return 0
	}

	build, err := buildLister.Builds(image.Namespace).Get(image.Status.LatestBuildRef)
	if err != nil {
		return 0
	}

	return len(build.Status.StepsCompleted)
}

func status(image *v1alpha1.Image) string {
	condition := image.Status.GetCondition(duckv1alpha1.ConditionReady)
	if condition == nil {
		return string(v1.ConditionUnknown)
	}

	return string(condition.Status)
}
