package main

import (
	"encoding/json"
	"flag"
	"github.com/matthewmcnew/build-service-visualization/images"
	"github.com/pivotal/kpack/pkg/client/clientset/versioned"
	"github.com/pivotal/kpack/pkg/client/informers/externalversions"
	"github.com/pivotal/kpack/pkg/client/listers/build/v1alpha1"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"time"
)

var (
	kubeconfig = flag.String("kubeconfig", "/Users/matthewmcnew/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	masterURL  = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	clusterConfig, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	client, err := versioned.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("could not get Build client: %s", err)
	}

	informerFactory := externalversions.NewSharedInformerFactory(client, 10*time.Hour)
	imageInformer := informerFactory.Build().V1alpha1().Images()

	lister := imageInformer.Lister()

	stopChan := make(chan struct{})
	informerFactory.Start(stopChan)

	fs := http.FileServer(http.Dir("ui/build"))
	http.Handle("/", fs)

	http.Handle("/images", &imagesApi{lister})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type imagesApi struct {
	lister v1alpha1.ImageLister
}

func (a *imagesApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	images, err := images.Current(a.lister)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(images)
}
