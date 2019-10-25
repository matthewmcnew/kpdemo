package images

import (
	"errors"
	"github.com/apex/log"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	v1alpha1Listers "github.com/pivotal/kpack/pkg/client/listers/build/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
	"sync"
	"time"
)

type Image struct {
	Name          string                         `json:"name"`
	Namespace     string                         `json:"namespace"`
	BuildCount    int64                          `json:"buildCount"`
	Status        string                         `json:"status"`
	BuildMetadata v1alpha1.BuildpackMetadataList `json:"buildMetadata"`
	Completed     int                            `json:"completed"`
	Remaining     int                            `json:"remaining"`
	CreatedAt     time.Time                      `json:"createdAt"`
	Tag           string                         `json:"tag"`
	LatestImage   string                         `json:"latestImage"`
	RunImage      string                         `json:"runImage"`
}

func Current(lister v1alpha1Listers.ImageLister, buildLister v1alpha1Listers.BuildLister) ([]Image, error) {
	kpackImages, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	images := make([]Image, 0, len(kpackImages))
	for _, i := range kpackImages {

		lastCompletedBuild, err := lastCompletedBuild(buildLister, i)
		if err != nil {
			log.Info(err.Error())
			continue
		}

		done, remaining := remaining(buildLister, i)
		images = append(images, Image{
			Name:          i.Name,
			Tag:           i.Spec.Tag,
			LatestImage:   i.Status.LatestImage,
			Namespace:     i.Namespace,
			BuildCount:    i.Status.BuildCounter,
			Status:        status(i),
			CreatedAt:     i.CreationTimestamp.Time,
			Completed:     done,
			Remaining:     remaining,
			BuildMetadata: lastCompletedBuild.Status.BuildMetadata,
			RunImage:      lastCompletedBuild.Status.Stack.RunImage,
		})
	}

	return images, nil
}

func lastCompletedBuild(buildLister v1alpha1Listers.BuildLister, image *v1alpha1.Image) (*v1alpha1.Build, error) {
	buildRef := image.Status.LatestBuildRef
	if buildRef == "" {
		return nil, errors.New("build not ready yet :)")
	}

	key := image.Name + "-" + image.Namespace
	if image.Status.LatestImage != "" && image.Status.GetCondition(duckv1alpha1.ConditionReady).IsUnknown() {
		var ok bool
		buildRef, ok = cache[key]
		if !ok {
			return nil, errors.New("coulding find cache key for build")
		}
	}

	build, err := buildLister.Builds(image.Namespace).Get(buildRef)
	if err != nil {
		return nil, err
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cache[key] = buildRef

	return build, nil
}

var cacheMutex = &sync.Mutex{}
var cache = map[string]string{}

func remaining(buildLister v1alpha1Listers.BuildLister, image *v1alpha1.Image) (int, int) {
	if image.Status.LatestBuildRef == "" {
		return 0, 10
	}

	if image.Status.GetCondition(duckv1alpha1.ConditionReady).IsTrue() {
		return 1, 1
	}

	if image.Status.GetCondition(duckv1alpha1.ConditionReady).IsFalse() {
		return 1, 1
	}

	//todo short circuit

	build, err := buildLister.Builds(image.Namespace).Get(image.Status.LatestBuildRef)
	if err != nil {
		return 0, -1
	}

	if len(build.Status.StepStates) == 0 {
		return 0, -1
	}

	return len(build.Status.StepsCompleted), len(build.Status.StepStates)
}

func status(image *v1alpha1.Image) string {
	condition := image.Status.GetCondition(duckv1alpha1.ConditionReady)
	if condition == nil {
		return string(v1.ConditionUnknown)
	}

	return string(condition.Status)
}
