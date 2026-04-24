package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/cluster"
	"github.com/openshift/must-gather-logging/internal/gather/collection"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	"github.com/openshift/must-gather-logging/internal/gather/console"
	"github.com/openshift/must-gather-logging/internal/gather/monitoring"
	"github.com/openshift/must-gather-logging/internal/gather/storage"
	"github.com/openshift/must-gather-logging/internal/utils"
)

const ()

var (
	// BaseNamespaces are the core namespaces to always inspect
	BaseNamespaces = []string{
		"openshift-operator-lifecycle-manager",
		"openshift-operators-redhat",
		"openshift-operators",
		common.DefaultNamespace,
	}
)

func main() {
	var (
		basePath string
		cacheDir string
	)

	flag.StringVar(&basePath, "base-path", "", "Base collection path (required)")
	flag.StringVar(&cacheDir, "cache-dir", "", "Cache directory for oc commands")
	flag.Parse()

	if basePath == "" {
		log.Fatal("Error: --base-path is required")
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		log.Fatalf("Failed to create base directory: %v", err)
	}

	if cacheDir == "" {
		cacheDir = path.Join(basePath, ".cache")
	}

	// Set up debug log file
	logFilePath := path.Join(basePath, "gather-debug.log")
	if err := utils.SetLogFile(logFilePath); err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer utils.CloseLogFile()

	utils.Log("must-gather logs are located at: '%s'", logFilePath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client := oc.NewClient(ctx, basePath, cacheDir)

	allNamespaces := mapset.NewSet(BaseNamespaces...)
	if exists, err := common.HasCrd(client, collection.KindClusterLogForwarder); exists {
		// Discover all namespaces that have ClusterLogForwarders
		clfNamespaces, err := common.GetResourceNamespaces(client, collection.KindClusterLogForwarder)
		if err != nil {
			log.Fatalf("failed to discover %q namespaces: %w", collection.KindClusterLogForwarder, err)
		}

		allNamespaces = allNamespaces.Union(clfNamespaces)
		// Only gather collection resources from namespaces that have ClusterLogForwarders
		collection.GatherResources(client, clfNamespaces)
	} else {
		log.Fatalf("failed check for crd %q: %v", collection.KindClusterLogForwarder, err)
	}
	nsList := allNamespaces.ToSlice()
	if err := cluster.GatherResources(client, nsList); err != nil {
		utils.Log("Failed to gather cluster resources: %v", err)
	}

	if err := console.GatherUIPlugin(client); err != nil {
		utils.Log("Warning: failed to gather UIPlugin: %v", err)
	}

	if err := storage.GatherResources(client, common.DefaultNamespace); err != nil {
		utils.Log("Failed to gather storage resources: %v", err)
	}

	if err := monitoring.GatherResources(client); err != nil {
		utils.Log("Failed to gather monitoring resources: %v", err)
	}

	log.Println("All resources gathered successfully")
}
