package server

import (
	"encoding/json"
	"github.com/matthewmcnew/build-service-visualization/images"
	"github.com/matthewmcnew/build-service-visualization/k8s"
	_ "github.com/matthewmcnew/build-service-visualization/statik"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/client/informers/externalversions"
	"github.com/pivotal/kpack/pkg/client/listers/build/v1alpha1"
	"github.com/rakyll/statik/fs"
	"log"
	"net/http"
	"time"
)

func Serve() {
	clusterConfig, err := k8s.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("could not get Build client: %s", err)
	}

	informerFactory := externalversions.NewSharedInformerFactory(client, 10*time.Hour)
	imageInformer := informerFactory.Build().V1alpha1().Images()
	buildInformer := informerFactory.Build().V1alpha1().Builds()

	imageLister := imageInformer.Lister()
	buildLister := buildInformer.Lister()

	stopChan := make(chan struct{})
	informerFactory.Start(stopChan)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	//fs := http.FileServer(http.Dir("ui/build"))
	http.Handle("/", http.FileServer(statikFS))

	http.Handle("/images", &imagesApi{imageLister: imageLister, buildLister: buildLister})
	log.Fatal(http.ListenAndServe(":8081", nil))
}

type imagesApi struct {
	imageLister v1alpha1.ImageLister
	buildLister v1alpha1.BuildLister
}

func (a *imagesApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	images, err := images.Current(a.imageLister, a.buildLister)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(images)
}
