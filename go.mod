module github.com/matthewmcnew/build-service-visualization

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/apex/log v1.1.2-0.20190827100214-baa5455d1012
	github.com/buildpacks/imgutil v0.0.0-20200115144305-2289fbc194a6
	github.com/google/go-containerregistry v0.0.0-20191018211754-b77a90c667af
	github.com/goombaio/namegenerator v0.0.0-20181006234301-989e774b106e
	github.com/pivotal/kpack v0.0.6
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.5
	k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	knative.dev/pkg v0.0.0-20191107185656-884d50f09454
)
