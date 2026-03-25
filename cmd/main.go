package main

import (
	"context"
	"flag"
	"log"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/cluster"
	"github.com/openshift/must-gather-logging/internal/gather/collection"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	"github.com/openshift/must-gather-logging/internal/gather/monitoring"
	"github.com/openshift/must-gather-logging/internal/gather/storage"
)

const ()

var (
	// BaseNamespaces are the core namespaces to always inspect
	BaseNamespaces = []string{
		"openshift-operator-lifecycle-manager",
		"openshift-operators-redhat",
		"openshift-operators",
		collection.DefaultNamespace,
	}
)

func main() {
	var (
		basePath  string
		namespace string
		cacheDir  string
	)

	flag.StringVar(&basePath, "base-path", "", "Base collection path (required)")
	//flag.StringVar(&namespace, "namespace", defaultNamespace, "Namespace to inspect")
	flag.StringVar(&cacheDir, "cache-dir", "", "Cache directory for oc commands")
	flag.Parse()

	if basePath == "" {
		log.Fatal("Error: --base-path is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client := oc.NewClient(ctx, basePath, cacheDir)

	namespaces := mapset.NewSet(BaseNamespaces...)
	if exists, err := common.HasCrd(client, collection.KindClusterLogForwarder); exists {
		// Discover all namespaces to inspect
		clfNamespaces, err := common.GetResourceNamespaces(client, collection.KindClusterLogForwarder)
		if err != nil {
			log.Fatalf("failed to discover %q namespaces: %w", collection.KindClusterLogForwarder, err)
		}

		namespaces = namespaces.Union(clfNamespaces)
		collection.GatherResources(client, namespaces)
	} else {
		log.Fatalf("failed check for crd %q: %v", collection.KindClusterLogForwarder, err)
	}
	nsList := namespaces.ToSlice()
	if err := cluster.GatherResources(client, nsList); err != nil {
		log.Fatalf("Failed to gather cluster resources: %v", err)
	}

	if err := storage.GatherResources(client, namespace); err != nil {
		log.Fatalf("Failed to gather storage resources: %v", err)
	}

	if err := monitoring.GatherResources(client); err != nil {
		log.Fatalf("Failed to gather monitoring resources: %v", err)
	}

	log.Println("All resources gathered successfully")
}
