package populate

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/buildpack/imgutil"
	"github.com/buildpack/imgutil/remote"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
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

	runImage := registry + "/run:pbdemo"
	runImg, err := remote.NewImage(runImage, authn.DefaultKeychain, remote.FromBaseImage("cloudfoundry/run:base-cnb"))
	if err != nil {
		return Relocated{}, err
	}

	err = runImg.Save()
	if err != nil {
		return Relocated{}, err
	}

	builderImage := registry + "/builder:pbdemo"
	builderImg, err := remote.NewImage(builderImage, authn.DefaultKeychain, remote.FromBaseImage("cloudfoundry/cnb:bionic"))
	if err != nil {
		return Relocated{}, err
	}

	layer, err := stackLayer(tmpDir, runImage)
	if err != nil {
		return Relocated{}, err
	}

	err = builderImg.AddLayer(layer)
	if err != nil {
		return Relocated{}, err
	}

	var md = map[string]interface{}{}
	if ok, err := getLabel(builderImg, metadataLabel, &md); err != nil {
		return Relocated{}, err
	} else if !ok {
		return Relocated{}, fmt.Errorf("builder %s missing label %s -- try recreating builder", builderImg.Name())
	}

	stack, ok := md["stack"].(map[string]interface{})
	if !ok {
		return Relocated{}, fmt.Errorf("builder Image does not have stack metadata")
	}

	runImageMetadata, ok := stack["runImage"].(map[string]interface{})
	if !ok {
		return Relocated{}, fmt.Errorf("builder Image does not have run image metadata")
	}
	runImageMetadata["image"] = runImage
	runImageMetadata["mirrors"] = nil

	err = setLabel(builderImg, metadataLabel, md)
	if err != nil {
		return Relocated{}, err
	}

	return Relocated{
		BuilderImage: builderImage,
		RunImage:     runImage,
	}, builderImg.Save()
}

func stackLayer(dest, runImage string) (string, error) {
	buf := &bytes.Buffer{}
	err := toml.NewEncoder(buf).Encode(StackMetadata{RunImage: RunImageMetadata{
		Image: runImage,
	}})
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal stack.toml")
	}

	layerTar := filepath.Join(dest, "stack.tar")
	err = CreateSingleFileTar(layerTar, "/cnb/stack.toml", buf.String())
	if err != nil {
		return "", errors.Wrapf(err, "failed to create stack.toml layer tar")
	}

	return layerTar, nil
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

func setLabel(image imgutil.Image, label string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "marshalling data to JSON for label %s", label)
	}
	if err := image.SetLabel(label, string(dataBytes)); err != nil {
		return errors.Wrapf(err, "setting label %s", label)
	}
	return nil
}

func getLabel(image imgutil.Image, label string, obj interface{}) (ok bool, err error) {
	labelData, err := image.Label(label)
	if err != nil {
		return false, errors.Wrapf(err, "retrieving label %s", label)
	}
	if labelData != "" {
		if err := json.Unmarshal([]byte(labelData), obj); err != nil {
			return false, errors.Wrapf(err, "unmarshalling label %s", label)
		}
		return true, nil
	}
	return false, nil
}

type StackMetadata struct {
	RunImage RunImageMetadata `json:"runImage" toml:"run-image"`
}

type RunImageMetadata struct {
	Image   string   `json:"image" toml:"image"`
	Mirrors []string `json:"mirrors" toml:"mirrors"`
}
