package main

import (
	"encoding/base64"
	"flag"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/goombaio/namegenerator"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	kubeconfig = flag.String("kubeconfig", "/Users/matthewmcnew/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	masterURL  = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	count      = flag.String("count", "20", "number of images to create")
)

func main() {
	flag.Parse()

	clusterConfig, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	k8sclient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("could not get Build client: %s", err)
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("could not get Build client: %s", err)
	}

	c := loadConfig()

	const namespace = "demo-team"
	_, err = k8sclient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Fatalf(err.Error())
	}

	secret, err := k8sclient.CoreV1().Secrets(namespace).Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "dockersecret",
			Annotations: map[string]string{
				"build.pivotal.io/docker": c.registry,
			},
		},
		StringData: map[string]string{
			"username": c.username,
			"password": c.password,
		},
		Type: v1.SecretTypeBasicAuth,
	})
	noError(err)

	serviceAccount, err := k8sclient.CoreV1().ServiceAccounts(namespace).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "serviceaccount",
		},
		Secrets: []v1.ObjectReference{
			{
				Name: secret.Name,
			},
		},
	})
	noError(err)

	const builderName = "demo-builder"
	_, err = client.BuildV1alpha1().ClusterBuilders().Create(&v1alpha1.ClusterBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name: builderName,
		},
		Spec: v1alpha1.BuilderSpec{
			Image:        "registry.default.svc.cluster.local:5000/builder",
			UpdatePolicy: v1alpha1.Polling,
		},
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Fatalf(err.Error())
	}

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	cache := resource.MustParse("100Mi")
	for i := 1; i <= c.count; i++ {
		image, err := client.BuildV1alpha1().Images(namespace).Create(&v1alpha1.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name: nameGenerator.Generate(),
			},
			Spec: v1alpha1.ImageSpec{
				Tag: c.imageTag,
				Builder: v1alpha1.ImageBuilder{
					TypeMeta: metav1.TypeMeta{
						Kind: "ClusterBuilder",
					},
					Name: builderName,
				},
				ServiceAccount: serviceAccount.Name,
				Source: v1alpha1.SourceConfig{
					Git: &v1alpha1.Git{
						URL:      "https://github.com/matthewmcnew/sample-java-app",
						Revision: "dbba68cee6473b5df51a1a43806d920d2ed4e4ee",
					},
				},
				CacheSize:                &cache,
				FailedBuildHistoryLimit:  nil,
				SuccessBuildHistoryLimit: nil,
				ImageTaggingStrategy:     v1alpha1.None,
				Build:                    v1alpha1.ImageBuild{},
			},
		})
		if err != nil && !errors.IsAlreadyExists(err) {
			log.Fatalf(err.Error())
		}

		log.Printf("created image %s", image.Name)
		time.Sleep(3 * time.Second)
	}

}

func noError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

type config struct {
	builder      string
	testRegistry string
	imageTag     string
	username     string
	password     string
	registry     string
	count        int
}

func loadConfig() config {
	registry, found := os.LookupEnv("IMAGE_REGISTRY")
	if !found {
		log.Fatal("IMAGE_REGISTRY env is needed for population")
	}

	imageTag := registryTag(registry)

	reg, err := name.ParseReference(registry, name.WeakValidation)
	if err != nil {
		log.Fatalf("Could not parse %s", imageTag)
	}

	auth, err := authn.DefaultKeychain.Resolve(reg.Context().Registry)
	if err != nil {
		log.Fatalf("Could not find keychain for%s", imageTag)
	}

	basicAuth, err := auth.Authorization()
	if err != nil {
		log.Fatalf("Could not get auth for%s", imageTag)
	}

	username, password, ok := parseBasicAuth(basicAuth)
	if !ok {
		log.Fatal("could not parse auth")
	}

	parsedCount, err := strconv.ParseInt(*count, 10, 64)
	if err != nil {
		log.Fatalf("Could not parse cout: %s", *count)
	}

	return config{
		testRegistry: registry,
		username:     username,
		password:     password,
		count:        int(parsedCount),
		imageTag:     imageTag,
		registry:     reg.Context().RegistryStr(),
	}
}

func registryTag(registry string) string {
	return registry + "/kpack-demo"
}

// net/http request.go
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}
