package collection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	"github.com/openshift/must-gather-logging/internal/utils"
)

// GatherOperatorResources gathers cluster logging operator resources
func GatherOperatorResources(client *oc.Client, namespace string) error {
	utils.Log("BEGIN <gather_cluster_logging_operator_resources> from namespace: %s", namespace)

	cloFolder := filepath.Join(client.BasePath, "cluster-logging", "clo")
	utils.Log("Creating namespace directory: %s", cloFolder)
	if err := os.MkdirAll(cloFolder, 0755); err != nil {
		return fmt.Errorf("failed to create clo folder: %w", err)
	}

	if err := gatherOperatorPodEnv(client, namespace, cloFolder); err != nil {
		utils.Log("Warning: failed to gather operator pod environment: %v", err)
	}

	// Gather version from CSV
	if err := gatherOperatorVersion(client, namespace, cloFolder); err != nil {
		utils.Log("Warning: failed to gather operator version: %v", err)
	}

	utils.Log("END <gather_cluster_logging_operator_resources> from namespace: %s", namespace)
	return nil
}

// gatherOperatorPodEnv gathers environment information from cluster-logging-operator pods
func gatherOperatorPodEnv(client *oc.Client, namespace, outputDir string) error {
	utils.Log("Gathering data for 'cluster-logging-operator' from namespace: %s", namespace)

	// Get pods with label name=cluster-logging-operator
	podList, err := common.GetPodsBySelector(client, namespace, "name=cluster-logging-operator")
	if err != nil {
		return fmt.Errorf("failed to get operator pods: %w", err)
	}

	if len(podList) == 0 {
		utils.Log("No cluster-logging-operator pods found")
		return nil
	}

	for _, pod := range podList {
		if pod == "" {
			continue
		}

		utils.Log("Inspecting %s", pod)
		if err := common.GetEnv(client, pod, outputDir, namespace, "Dockerfile-.*operator*"); err != nil {
			utils.Log("Warning: failed to get env for pod %s: %v", pod, err)
		}
	}

	return nil
}

// gatherOperatorVersion gathers the operator version from CSV
func gatherOperatorVersion(client *oc.Client, namespace, outputDir string) error {
	utils.Log("Gathering 'version' from logging namespace: %s", namespace)

	// Get CSV names
	csvs, err := client.Get(oc.GetOptions{
		Resource:  "csv",
		Namespace: namespace,
		Output:    "name",
	})

	if err != nil {
		return fmt.Errorf("failed to get CSVs: %w", err)
	}

	// Find cluster-logging CSV
	csvList := utils.ParseLines(string(csvs))
	var csvName string
	for _, csv := range csvList {
		if strings.Contains(csv, "cluster-logging") || strings.Contains(csv, "clusterlogging") {
			csvName = csv
			break
		}
	}

	if csvName == "" {
		return fmt.Errorf("no cluster-logging CSV found")
	}

	// Get version info
	versionInfo, err := client.Get(oc.GetOptions{
		Resource:  csvName,
		Namespace: namespace,
		Output:    `jsonpath={.spec.displayName}{"/must-gather\n"}{.spec.version}`,
	})

	if err != nil {
		return fmt.Errorf("failed to get CSV version: %w", err)
	}

	versionFile := filepath.Join(outputDir, "version")
	return os.WriteFile(versionFile, versionInfo, 0644)
}
