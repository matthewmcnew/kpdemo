package buildpacks

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
)

func buildpackage(id, version, newVersion, source, destination string) (string, error) {
	reference, err := name.ParseReference(source)
	if err != nil {
		return "", err
	}

	sourceImage, err := remote.Image(reference, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", err
	}

	metadata := BuildpackLayerMetadata{}
	err = imagehelpers.GetLabel(sourceImage, "io.buildpacks.buildpack.layers", &metadata)
	if err != nil {
		return "", err
	}

	info, err := metadata.metadataFor(id, version)
	if err != nil {
		return "", err
	}

	hash, err := v1.NewHash(info.LayerDiffID)
	if err != nil {
		return "", err
	}

	bpl, err := sourceImage.LayerByDiffID(hash)
	if err != nil {
		return "", err
	}

	newBpL, err := rewriteLayer(bpl, version, newVersion)
	if err != nil {
		return "", err
	}

	newBuildpackage, err := random.Image(0, 0)
	if err != nil {
		return "", err
	}

	newBuildpackage, err = mutate.AppendLayers(newBuildpackage, newBpL)
	if err != nil {
		return "", err
	}

	newBuildpackage, err = imagehelpers.SetLabels(newBuildpackage, map[string]interface{}{
		"io.buildpacks.buildpack.layers": BuildpackLayerMetadata{
			id: {
				newVersion: info,
			},
		},
		"io.buildpacks.buildpackage.metadata": Metadata{
			BuildpackInfo: BuildpackInfo{
				Id:      id,
				Version: newVersion,
			},
			Stacks: info.Stacks,
		},
	})
	if err != nil {
		return "", err
	}

	reference, err = name.ParseReference(destination)
	if err != nil {
		return "", err
	}

	err = remote.Write(reference, newBuildpackage, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", err
	}

	digest, err := newBuildpackage.Digest()
	if err != nil {
		return "", err
	}

	identifer := fmt.Sprintf("%s@%s", destination, digest.String())

	fmt.Printf("successfully wrote %s@%s to %s\n", id, newVersion, identifer)
	return identifer, nil
}

type BuildpackLayerMetadata map[string]map[string]BuildpackLayerInfo

func (m BuildpackLayerMetadata) metadataFor(id string, version string) (BuildpackLayerInfo, error) {
	bps, ok := m[id]
	if !ok {
		var available []string

		for bp := range m {
			available = append(available, bp)
		}

		return BuildpackLayerInfo{}, errors.Errorf("could not find %s, options: %s", id, available)
	}

	info, ok := bps[version]
	if !ok {
		var available []string

		for v := range bps {
			available = append(available, fmt.Sprintf("%s@%s", id, v))
		}

		return BuildpackLayerInfo{}, errors.Errorf("could not find %s@%s, options: %s", id, version, available)
	}

	return info, nil
}

type BuildpackLayerInfo struct {
	API         string  `json:"api"`
	Stacks      []Stack `json:"stacks,omitempty"`
	Order       Order   `json:"order,omitempty"`
	LayerDiffID string  `json:"layerDiffID"`
}

type Order []OrderEntry

type OrderEntry struct {
	Group []BuildpackRef `json:"group,omitempty"`
}

type BuildpackRef struct {
	BuildpackInfo `json:",inline"`
	Optional      bool `json:"optional,omitempty"`
}

type BuildpackInfo struct {
	Id      string `json:"id"`
	Version string `json:"version,omitempty"`
}

type Stack struct {
	ID     string   `json:"id"`
	Mixins []string `json:"mixins,omitempty"`
}

type Metadata struct {
	BuildpackInfo
	Stacks []Stack `toml:"stacks" json:"stacks,omitempty"`
}
