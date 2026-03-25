package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/collection"
	"github.com/openshift/must-gather-logging/internal/gather/storage"
)

const (
	defaultNamespace = "openshift-logging"
)

func main() {
	var (
		basePath  string
		namespace string
		cacheDir  string
	)

	flag.StringVar(&basePath, "base-path", "", "Base collection path (required)")
	flag.StringVar(&namespace, "namespace", defaultNamespace, "Namespace to inspect")
	flag.StringVar(&cacheDir, "cache-dir", "", "Cache directory for oc commands")
	flag.Parse()

	if basePath == "" {
		log.Fatal("Error: --base-path is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client := oc.NewClient(ctx, basePath, cacheDir)

	if err := collection.GatherResources(client, namespace); err != nil {
		log.Fatalf("Failed to gather collection resources: %v", err)
	}

	if err := collection.GatherOperatorResources(client, namespace); err != nil {
		log.Fatalf("Failed to gather operator resources: %v", err)
	}

	if err := storage.GatherResources(client, namespace); err != nil {
		log.Fatalf("Failed to gather storage resources: %v", err)
	}

	log.Println("All resources gathered successfully")
}
