module github.com/matthewmcnew/kpdemo

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/apex/log v1.3.0
	github.com/fatih/color v1.9.0
	github.com/google/go-containerregistry v0.1.1
	github.com/goombaio/namegenerator v0.0.0-20181006234301-989e774b106e
	github.com/pivotal/kpack v0.1.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace (
	k8s.io/client-go => k8s.io/client-go v0.17.6
)
