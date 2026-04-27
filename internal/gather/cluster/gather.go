package cluster

import (
	"strings"
	"sync"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/utils/log"
)

var (
	// clusterResources are cluster-scoped resources to inspect
	clusterResources = []string{
		"clusterroles",
		"clusterrolebindings",
		"clusterversion",
		"customresourcedefinitions",
		"persistentvolumes",
		"machineconfigpool",
		"nodes",
	}

	// namespacedResources are namespace-scoped resources to inspect
	namespacedResources = []string{
		"clusterserviceversions",
		"configmaps",
		"events",
		"installplans",
		"logfilemetricexporter",
		"pods",
		"roles",
		"rolebindings",
		"serviceaccounts",
		"subscriptions",
	}
)

// GatherResources gathers cluster-wide resources
func GatherResources(client *oc.Client, namespaces []string, redactSecrets bool) error {
	log.Begin(1, "inspecting cluster resources and namespaces...")

	// Inspect cluster-scoped resources
	if err := inspectClusterResources(client); err != nil {
		log.Warn("failed to inspect cluster resources: %v", err)
	}

	// Inspect namespaces
	if err := inspectNamespaces(client, namespaces); err != nil {
		log.Warn("failed to inspect namespaces: %v", err)
	}

	// Inspect namespace-scoped resources
	if err := inspectNamespacedResources(client, namespaces); err != nil {
		log.Warn("failed to inspect namespaced resources: %v", err)
	}

	// Redact secrets if requested
	if redactSecrets {
		if err := redactSecretsInNamespaces(client.BasePath, namespaces); err != nil {
			log.Warn("failed to redact secrets: %v", err)
		}
	}

	log.End(1, "inspecting cluster resources...")
	return nil
}

// inspectClusterResources inspects cluster-scoped resources in parallel
func inspectClusterResources(client *oc.Client) error {
	adm := client.Adm()
	var wg sync.WaitGroup

	for _, resource := range clusterResources {
		wg.Add(1)
		go func(res string) {
			defer wg.Done()
			log.Begin(2, "inspecting cluster resource %s ...", res)
			if err := adm.Inspect(oc.InspectOptions{
				Resources: []string{res},
			}); err != nil {
				log.Warn("failed to inspect cluster resource %s: %v", res, err)
			}
		}(resource)
	}

	wg.Wait()
	return nil
}

// inspectNamespaces inspects namespace objects in parallel
func inspectNamespaces(client *oc.Client, namespaces []string) error {
	adm := client.Adm()
	var wg sync.WaitGroup

	for _, ns := range namespaces {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()
			log.Begin(2, "inspecting namespace %s ...", namespace)
			if err := adm.Inspect(oc.InspectOptions{
				Resources: []string{"ns/" + namespace},
			}); err != nil {
				log.Warn("failed to inspect namespace %s: %v", namespace, err)
			}
		}(ns)
	}

	wg.Wait()
	return nil
}

// inspectNamespacedResources inspects namespace-scoped resources in parallel
func inspectNamespacedResources(client *oc.Client, namespaces []string) error {
	log.Begin(0, "inspecting namespaced resources ...")

	adm := client.Adm()
	resourceList := strings.Join(namespacedResources, ",")
	var wg sync.WaitGroup

	for _, ns := range namespaces {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()
			log.Begin(2, "inspecting %s/%s ...", namespace, resourceList)
			if err := adm.Inspect(oc.InspectOptions{
				Namespace: namespace,
				Resources: []string{resourceList},
			}); err != nil {
				log.Warn("failed to inspect namespaced resources in %s: %v", namespace, err)
			}
		}(ns)
	}

	wg.Wait()
	log.End(0, "inspecting namespaced resources ...")
	return nil
}
