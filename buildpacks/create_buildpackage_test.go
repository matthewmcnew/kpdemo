package buildpacks

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreate(t *testing.T) {
	_, err := buildpackage("org.cloudfoundry.go-compiler", "0.0.83", "cloudfoundry/cnb:bionic", "localhost:5001/rewrite")
	require.NoError(t, err)
}
