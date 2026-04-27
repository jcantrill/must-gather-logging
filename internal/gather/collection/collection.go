package collection

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	logger "github.com/openshift/must-gather-logging/internal/utils/log"
)

const (
	KindClusterLogForwarder = "clusterlogforwarders.observability.openshift.io"
)

func GatherResources(client *oc.Client, namespaces mapset.Set[string]) (err error) {
	if err = GatherOperatorResources(client, common.DefaultNamespace); err != nil {
		logger.Log("Failed to gather operator resources: %v", err)
	}
	namespaces.Each(func(ns string) bool {
		if err = GatherClusterLogForwarderResources(client, ns); err != nil {
			log.Fatalf("Failed to gather collection resources: %v", err)
		}
		return false
	})
	return nil
}

// GatherClusterLogForwarderResources gathers collection resources from a namespace
func GatherClusterLogForwarderResources(client *oc.Client, namespace string) (err error) {
	logger.Begin(0, "<gather_collection_resources> for namespace: %s", namespace)

	collectorFolder := filepath.Join(client.BasePath, "cluster-logging", "namespaces", namespace)
	if err := os.MkdirAll(collectorFolder, 0755); err != nil {
		return fmt.Errorf("failed to create clf folder: %w", err)
	}

	// Get ClusterLogForwarder.observability.openshift.io resources
	logger.Log("Exporting %s resources", KindClusterLogForwarder)

	names, err := common.GetResourceNames(client, KindClusterLogForwarder, namespace)
	if err != nil {
		return fmt.Errorf("failed to get clusterlogforwarders: %w", err)
	}
	if len(names) == 0 {
		logger.Log("No ClusterLogForwarders found in namespace")
		return nil
	}

	// Process each clf
	for _, clf := range names {
		if clf == "" {
			continue
		}

		if err = gatherCollectorData(client, namespace, clf, collectorFolder); err != nil {
			logger.Warn("failed to gather data for clf %s: %v", clf, err)
		}
	}

	logger.End(0, "<gather_collection_resources> for namespace: %s", namespace)
	return nil
}

func gatherCollectorData(client *oc.Client, namespace, collector, collectorFolder string) error {
	logger.Log("Gathering data for ClusterLogForwarder: %s", collector)

	// Inspect ClusterLogForwarders
	adm := client.Adm()
	if err := adm.Inspect(oc.InspectOptions{
		Namespace: namespace,
		Resources: []string{KindClusterLogForwarder},
	}); err != nil {
		logger.Warn("failed to inspect clusterlogforwarders: %v", err)
	}

	// Describe DaemonSet
	if err := describeDaemonSet(client, namespace, collector, collectorFolder); err != nil {
		logger.Warn("failed to describe daemonset: %v", err)
	}

	// Gather collector pods
	if err := gatherCollectorPods(client, namespace, collector, collectorFolder); err != nil {
		logger.Warn("failed to gather collector pods: %v", err)
	}

	// Gather vector.toml from configmap
	if err := gatherVectorConfig(client, namespace, collector, collectorFolder); err != nil {
		logger.Warn("failed to gather vector config: %v", err)
	}

	return nil
}

func describeDaemonSet(client *oc.Client, namespace, collector, outputDir string) error {
	logger.Log("Describe DaemonSet ds/%s", collector)

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

func gatherCollectorPods(client *oc.Client, namespace, collector, outputDir string) (err error) {
	logger.Log("Gathering collector pods")

	// Get pods with the collector labels
	selector := fmt.Sprintf("app.kubernetes.io/instance=%s,app.kubernetes.io/component=collector", collector)

	podList, err := common.GetPodsBySelector(client, namespace, selector)
	if err != nil {
		return err
	}
	if len(podList) == 0 {
		logger.Log("No collector pods found")
		return nil
	}

	// Describe each pod
	for _, pod := range podList {
		if pod == "" {
			continue
		}

		logger.Log("Describe collector pod: %s", pod)

		var output []byte
		output, err = client.Describe(oc.DescribeOptions{
			Resource:  "pod/" + pod,
			Namespace: namespace,
		})

		if err != nil {
			logger.Warn("failed to describe pod %s: %v", pod, err)
			continue
		}

		podFile := filepath.Join(outputDir, pod+".describe")
		if err = os.WriteFile(podFile, output, 0644); err != nil {
			logger.Warn("failed to write pod describe file: %v", err)
		}
	}

	return nil
}

func gatherVectorConfig(client *oc.Client, namespace, collector, outputDir string) error {
	configName := collector + "-config"
	logger.Log("Gathering %s#vector.toml from namespace: %s", configName, namespace)

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
		logger.Log("No vector.toml found in configmap %s", configName)
		return nil
	}

	tomlFile := filepath.Join(outputDir, fmt.Sprintf("configmap_%s_vector.toml", configName))
	return os.WriteFile(tomlFile, vectorToml, 0644)
}
