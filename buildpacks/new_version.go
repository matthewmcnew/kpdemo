package buildpacks

import (
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

func newVersion(id, v string) (string, error) {
	version, err := semver.NewVersion(v)
	if err != nil {
		return "", errors.Errorf("could not calculate next version for %s@%s. Is it valid semver?", id, version)
	}

	return version.IncPatch().String(), nil
}
