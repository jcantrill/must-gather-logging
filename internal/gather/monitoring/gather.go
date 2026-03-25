package monitoring

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	"github.com/openshift/must-gather-logging/internal/utils"
)

const (
	monitoringNamespace = "openshift-monitoring"
	prometheusLabel     = "prometheus=k8s"
)

// GatherResources gathers monitoring and prometheus resources
func GatherResources(client *oc.Client) error {
	utils.Log("BEGIN gathering alerts ...")

	monitoringPath := filepath.Join(client.BasePath, "monitoring")
	if err := os.MkdirAll(monitoringPath, 0755); err != nil {
		return fmt.Errorf("failed to create monitoring directory: %w", err)
	}

	// Find prometheus pods
	promPods, err := getPrometheusPods(client)
	if err != nil {
		return fmt.Errorf("failed to get prometheus pods: %w", err)
	}

	if len(promPods) == 0 {
		utils.Log("No prometheus pods found")
		return nil
	}

	utils.Log("INFO: Found %d replicas - %v", len(promPods), promPods)

	// Get first ready prometheus pod
	readyPod, err := getFirstReadyPrometheusPod(client)
	if err != nil {
		return fmt.Errorf("failed to get ready prometheus pod: %w", err)
	}

	if readyPod == "" {
		utils.Log("No ready prometheus pods found")
		return nil
	}

	// Gather prometheus rules
	if err := gatherPrometheusRules(client, readyPod, monitoringPath); err != nil {
		utils.Log("Warning: failed to gather prometheus rules: %v", err)
	}

	utils.Log("END gathering alerts ...")
	return nil
}

// getPrometheusPods returns all prometheus pods in openshift-monitoring namespace
func getPrometheusPods(client *oc.Client) ([]string, error) {
	return common.GetPodsBySelector(client, monitoringNamespace, prometheusLabel)
}

// getFirstReadyPrometheusPod returns the first running prometheus pod
func getFirstReadyPrometheusPod(client *oc.Client) (string, error) {
	pods, err := common.GetPodsBySelectorWithFieldSelector(client, monitoringNamespace, prometheusLabel, "status.phase==Running")
	if err != nil {
		return "", err
	}

	if len(pods) == 0 {
		return "", nil
	}

	return pods[0], nil
}

// gatherPrometheusRules queries prometheus for alert rules and saves the output
func gatherPrometheusRules(client *oc.Client, pod, monitoringPath string) error {
	utils.Log("INFO: Getting rules from %s", pod)

	prometheusPath := filepath.Join(monitoringPath, "prometheus")
	if err := os.MkdirAll(prometheusPath, 0755); err != nil {
		return fmt.Errorf("failed to create prometheus directory: %w", err)
	}

	// Execute curl inside the prometheus pod to query the API
	output, err := client.Exec(oc.ExecOptions{
		Pod:       pod,
		Namespace: monitoringNamespace,
		Container: "prometheus",
		Command:   []string{"/bin/bash", "-c", "curl -sG http://localhost:9090/api/v1/rules"},
	})

	rulesFile := filepath.Join(prometheusPath, "rules.json")
	stderrFile := filepath.Join(prometheusPath, "rules.stderr")

	if err != nil {
		// Write error to stderr file
		errMsg := fmt.Sprintf("Failed to query prometheus: %v\n", err)
		os.WriteFile(stderrFile, []byte(errMsg), 0644)
		return err
	}

	// Write successful output to json file
	if err := os.WriteFile(rulesFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	// Create empty stderr file to match bash script behavior
	os.WriteFile(stderrFile, []byte{}, 0644)

	return nil
}
