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

func buildpackage(id, version, source, destination string) (string, error) {
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

	info, layers, err := metadata.metadataAndLayersFor(BuildpackLayerMetadata{}, sourceImage, id, version)
	if err != nil {
		return "", err
	}

	newBuildpackage, err := random.Image(0, 0)
	if err != nil {
		return "", err
	}

	newBuildpackage, err = mutate.AppendLayers(newBuildpackage, layers...)
	if err != nil {
		return "", err
	}

	newBuildpackage, err = imagehelpers.SetLabels(newBuildpackage, map[string]interface{}{
		"io.buildpacks.buildpack.layers": info,
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

	return identifer, nil
}

type BuildpackLayerMetadata map[string]map[string]BuildpackLayerInfo

func (m BuildpackLayerMetadata) metadataAndLayersFor(initalMetadata BuildpackLayerMetadata, sourceImage v1.Image, id string, version string) (BuildpackLayerMetadata, []v1.Layer, error) {
	bps, ok := m[id]
	if !ok {
		return m, nil, errors.Errorf("could not find %s", id)
	}

	info, ok := bps[version]
	if !ok {
		return m, nil, errors.Errorf("could not find %s@%s", id, version)
	}

	var layers []v1.Layer
	var newOrder Order
	for _, oe := range info.Order {
		var newGroup []BuildpackRef
		for _, g := range oe.Group {
			var err error
			var ls []v1.Layer

			initalMetadata, ls, err = m.metadataAndLayersFor(initalMetadata, sourceImage, g.Id, g.Version)
			if err != nil {
				return m, nil, err
			}
			layers = append(layers, ls...)

			newVersion, err := newVersion(g.Id, g.Version)
			if err != nil {
				return m, nil, err
			}

			newGroup = append(newGroup, BuildpackRef{
				BuildpackInfo: BuildpackInfo{
					Id:      g.Id,
					Version: newVersion,
				},
				Optional: g.Optional,
			})
		}
		newOrder = append(newOrder, OrderEntry{Group: newGroup})
	}

	hash, err := v1.NewHash(info.LayerDiffID)
	if err != nil {
		return m, nil, err
	}

	bpl, err := sourceImage.LayerByDiffID(hash)
	if err != nil {
		return m, nil, err
	}

	newVersion, err := newVersion(id, version)
	if err != nil {
		return m, nil, err
	}
	newBpL, err := rewriteLayer(bpl, version, newVersion)
	if err != nil {
		return m, nil, err
	}

	diffID, err := newBpL.DiffID()
	if err != nil {
		return m, nil, err
	}

	_, ok = initalMetadata[id]
	if !ok {
		initalMetadata[id] = map[string]BuildpackLayerInfo{}
	}
	initalMetadata[id][newVersion] = info
	initalMetadata[id][newVersion] = BuildpackLayerInfo{
		API:         info.API,
		Stacks:      info.Stacks,
		Order:       newOrder,
		LayerDiffID: diffID.String(),
	}

	return initalMetadata, append(layers, newBpL), nil
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
