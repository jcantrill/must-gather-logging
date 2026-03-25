package collection

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	"github.com/openshift/must-gather-logging/internal/utils"
)

// GatherResources gathers collection resources from a namespace
func GatherResources(client *oc.Client, namespace string) error {
	utils.Log("BEGIN <gather_collection_resources> for namespace: %s", namespace)

	collectorFolder := filepath.Join(client.BasePath, "cluster-logging", "namespaces", namespace)
	if err := os.MkdirAll(collectorFolder, 0755); err != nil {
		return fmt.Errorf("failed to create collector folder: %w", err)
	}

	// Get ClusterLogForwarder.observability.openshift.io resources
	utils.Log("Exporting ClusterLogForwarder.observability.openshift.io resources")

	clfs, err := client.Get(oc.GetOptions{
		Resource:       "clusterlogforwarder.observability.openshift.io",
		Namespace:      namespace,
		Output:         common.ColumnsMetadataName,
		NoHeaders:      true,
		IgnoreNotFound: true,
	})

	if err != nil {
		return fmt.Errorf("failed to get clusterlogforwarders: %w", err)
	}

	collectors := utils.ParseLines(string(clfs))
	if len(collectors) == 0 {
		utils.Log("No ClusterLogForwarders found in namespace")
		return nil
	}

	// Process each collector
	for _, collector := range collectors {
		if collector == "" {
			continue
		}

		if err := gatherCollectorData(client, namespace, collector, collectorFolder); err != nil {
			utils.Log("Warning: failed to gather data for collector %s: %v", collector, err)
		}
	}

	utils.Log("END <gather_collection_resources> for namespace: %s", namespace)
	return nil
}

func gatherCollectorData(client *oc.Client, namespace, collector, collectorFolder string) error {
	utils.Log("Gathering data for ClusterLogForwarder: %s", collector)

	// Inspect ClusterLogForwarders
	adm := client.Adm()
	if err := adm.Inspect(oc.InspectOptions{
		Namespace: namespace,
		Resources: []string{"clusterlogforwarders.observability.openshift.io"},
	}); err != nil {
		utils.Log("Warning: failed to inspect clusterlogforwarders: %v", err)
	}

	// Describe DaemonSet
	if err := describeDaemonSet(client, namespace, collector, collectorFolder); err != nil {
		utils.Log("Warning: failed to describe daemonset: %v", err)
	}

	// Gather collector pods
	if err := gatherCollectorPods(client, namespace, collector, collectorFolder); err != nil {
		utils.Log("Warning: failed to gather collector pods: %v", err)
	}

	// Gather vector.toml from configmap
	if err := gatherVectorConfig(client, namespace, collector, collectorFolder); err != nil {
		utils.Log("Warning: failed to gather vector config: %v", err)
	}

	return nil
}

func describeDaemonSet(client *oc.Client, namespace, collector, outputDir string) error {
	utils.Log("Describe DaemonSet ds/%s", collector)

	output, err := client.Describe(oc.DescribeOptions{
		Resource:  "ds/" + collector,
		Namespace: namespace,
	})

	if err != nil {
		return err
	}

	descFile := filepath.Join(outputDir, collector+".describe")
	return os.WriteFile(descFile, output, 0644)
}

func gatherCollectorPods(client *oc.Client, namespace, collector, outputDir string) error {
	utils.Log("Gathering collector pods")

	// Get pods with the collector labels
	selector := fmt.Sprintf("app.kubernetes.io/instance=%s,app.kubernetes.io/component=collector", collector)

	podList, err := common.GetPodsBySelector(client, namespace, selector)
	if err != nil {
		return err
	}
	if len(podList) == 0 {
		utils.Log("No collector pods found")
		return nil
	}

	// Describe each pod
	for _, pod := range podList {
		if pod == "" {
			continue
		}

		utils.Log("Describe collector pod: %s", pod)

		output, err := client.Describe(oc.DescribeOptions{
			Resource:  "pod/" + pod,
			Namespace: namespace,
		})

		if err != nil {
			utils.Log("Warning: failed to describe pod %s: %v", pod, err)
			continue
		}

		podFile := filepath.Join(outputDir, pod+".describe")
		if err := os.WriteFile(podFile, output, 0644); err != nil {
			utils.Log("Warning: failed to write pod describe file: %v", err)
		}
	}

	return nil
}

func gatherVectorConfig(client *oc.Client, namespace, collector, outputDir string) error {
	configName := collector + "-config"
	utils.Log("Gathering %s#vector.toml from namespace: %s", configName, namespace)

	// Get vector.toml from configmap
	vectorToml, err := client.Get(oc.GetOptions{
		Resource:       "configmap/" + configName,
		Namespace:      namespace,
		Output:         "jsonpath={.data.vector\\.toml}",
		IgnoreNotFound: true,
	})

	if err != nil {
		return err
	}

	if len(vectorToml) == 0 {
		utils.Log("No vector.toml found in configmap %s", configName)
		return nil
	}

	tomlFile := filepath.Join(outputDir, fmt.Sprintf("configmap_%s_vector.toml", configName))
	return os.WriteFile(tomlFile, vectorToml, 0644)
}
