package images

import (
	"github.com/pivotal/kpack/pkg/client/listers/build/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
)

type Image struct {
	Ready bool
	Name  string
}

func Current(lister v1alpha1.ImageLister) ([]Image, error) {
	kpackImages, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	images := make([]Image, 0, len(kpackImages))
	for _, i := range kpackImages {
		images = append(images, Image{
			Name:  i.Name,
			Ready: i.Status.LatestImage != "",
		})
	}

	return images, nil
}
