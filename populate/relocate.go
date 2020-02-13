package populate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pivotal/kpack/pkg/registry/imagehelpers"
	"github.com/pkg/errors"
)

const metadataLabel = "io.buildpacks.builder.metadata"

type Relocated struct {
	BuilderImage string
	RunImage     string
}

func Relocate(registry string) (Relocated, error) {
	tmpDir, err := ioutil.TempDir("", "create-builder-scratch")
	if err != nil {
		return Relocated{}, err
	}
	defer os.RemoveAll(tmpDir)

	runRef, err := name.ParseReference("cloudfoundry/run:base-cnb")
	if err != nil {
		return Relocated{}, err
	}

	run, err := remote.Image(runRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return Relocated{}, err
	}

	relocatedRunRef, err := name.ParseReference(registry + "/pbdemo:run")
	if err != nil {
		return Relocated{}, err
	}

	_, err = save(relocatedRunRef, run)
	if err != nil {
		return Relocated{}, err
	}

	builderRef, err := name.ParseReference("cloudfoundry/cnb:bionic")
	if err != nil {
		return Relocated{}, err
	}

	builder, err := remote.Image(builderRef, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return Relocated{}, err
	}

	layer, err := stackLayer(tmpDir, relocatedRunRef.Name())
	if err != nil {
		return Relocated{}, err
	}

	builder, err = mutate.AppendLayers(builder, layer)
	if err != nil {
		return Relocated{}, err
	}

	var md = map[string]interface{}{}
	if err := imagehelpers.GetLabel(builder, metadataLabel, &md); err != nil {
		return Relocated{}, errors.Wrapf(err, "invalid label %s", metadataLabel)
	}

	stack, ok := md["stack"].(map[string]interface{})
	if !ok {
		return Relocated{}, fmt.Errorf("builder Image does not have stack metadata")
	}

	runImageMetadata, ok := stack["runImage"].(map[string]interface{})
	if !ok {
		return Relocated{}, fmt.Errorf("builder Image does not have run image metadata")
	}
	runImageMetadata["image"] = relocatedRunRef.Name()
	runImageMetadata["mirrors"] = nil

	builder, err = imagehelpers.SetLabels(builder, map[string]interface{}{
		metadataLabel: md,
	})
	if err != nil {
		return Relocated{}, err
	}

	relocatedBuilderRef, err := name.ParseReference(registry + "/pbdemo:builder")
	if err != nil {
		return Relocated{}, err
	}

	_, err = save(relocatedBuilderRef, builder)
	return Relocated{
		BuilderImage: relocatedBuilderRef.Name(),
		RunImage:     relocatedRunRef.Name(),
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

func stackLayer(dest, runImage string) (v1.Layer, error) {
	buf := &bytes.Buffer{}
	err := toml.NewEncoder(buf).Encode(StackMetadata{RunImage: RunImageMetadata{
		Image: runImage,
	}})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal stack.toml")
	}

	layerTar := filepath.Join(dest, "stack.tar")
	err = CreateSingleFileTar(layerTar, "/cnb/stack.toml", buf.String())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create stack.toml layer tar")
	}

	return tarball.LayerFromFile(layerTar, tarball.WithCompressionLevel(gzip.DefaultCompression))
}

func CreateSingleFileTar(tarFile, path, txt string) error {
	fh, err := os.Create(tarFile)
	if err != nil {
		return fmt.Errorf("create file for tar: %s", err)
	}
	defer fh.Close()

	tw := tar.NewWriter(fh)
	defer tw.Close()
	return AddFileToTar(tw, path, txt)
}

func AddFileToTar(tw *tar.Writer, path string, txt string) error {
	if err := tw.WriteHeader(&tar.Header{
		Name: path,
		Size: int64(len(txt)),
		Mode: 0644,
	}); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(txt)); err != nil {
		return err
	}
	return nil
}

type StackMetadata struct {
	RunImage RunImageMetadata `json:"runImage" toml:"run-image"`
}

type RunImageMetadata struct {
	Image   string   `json:"image" toml:"image"`
	Mirrors []string `json:"mirrors" toml:"mirrors"`
}
