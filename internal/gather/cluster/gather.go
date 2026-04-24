package cluster

import (
	"strings"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/utils"
)

var (
	// clusterResources are cluster-scoped resources to inspect
	clusterResources = []string{
		"nodes",
		"clusterroles",
		"clusterrolebindings",
		"persistentvolumes",
		"clusterversion",
		"machineconfigpool",
		"customresourcedefinitions",
	}

	// namespacedResources are namespace-scoped resources to inspect
	namespacedResources = []string{
		"pods",
		"roles",
		"rolebindings",
		"configmaps",
		"serviceaccounts",
		"events",
		"installplans",
		"subscriptions",
		"clusterserviceversions",
		"logfilemetricexporter",
	}
)

// GatherResources gathers cluster-wide resources
func GatherResources(client *oc.Client, namespaces []string) error {
	utils.Log("- BEGIN inspecting cluster resources and namespaces...")

	// Inspect cluster-scoped resources
	if err := inspectClusterResources(client); err != nil {
		utils.Log("Warning: failed to inspect cluster resources: %v", err)
	}

	// Inspect namespaces
	if err := inspectNamespaces(client, namespaces); err != nil {
		utils.Log("Warning: failed to inspect namespaces: %v", err)
	}

	// Inspect namespace-scoped resources
	if err := inspectNamespacedResources(client, namespaces); err != nil {
		utils.Log("Warning: failed to inspect namespaced resources: %v", err)
	}

	utils.Log("- END inspecting cluster resources...")
	return nil
}

// inspectClusterResources inspects cluster-scoped resources
func inspectClusterResources(client *oc.Client) error {
	adm := client.Adm()

	for _, resource := range clusterResources {
		utils.Log("-- BEGIN inspecting cluster resource %s ...", resource)
		if err := adm.Inspect(oc.InspectOptions{
			Resources: []string{resource},
		}); err != nil {
			utils.Log("Warning: failed to inspect cluster resource %s: %v", resource, err)
		}
	}

	return nil
}

// inspectNamespaces inspects namespace objects
func inspectNamespaces(client *oc.Client, namespaces []string) error {
	adm := client.Adm()

	for _, ns := range namespaces {
		utils.Log("-- BEGIN inspecting namespace %s ...", ns)
		if err := adm.Inspect(oc.InspectOptions{
			Resources: []string{"ns/" + ns},
		}); err != nil {
			utils.Log("Warning: failed to inspect namespace %s: %v", ns, err)
		}
	}

	return nil
}

// inspectNamespacedResources inspects namespace-scoped resources
func inspectNamespacedResources(client *oc.Client, namespaces []string) error {
	utils.Log("BEGIN inspecting namespaced resources ...")

	adm := client.Adm()

	for _, ns := range namespaces {
		resourceList := strings.Join(namespacedResources, ",")
		utils.Log("-- BEGIN inspecting %s/%s ...", ns, resourceList)
		if err := adm.Inspect(oc.InspectOptions{
			Namespace: ns,
			Resources: []string{resourceList},
		}); err != nil {
			utils.Log("Warning: failed to inspect namespaced resources in %s: %v", ns, err)
		}
	}

	utils.Log("END inspecting namespaced resources ...")
	return nil
}
