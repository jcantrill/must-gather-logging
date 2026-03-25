package storage

import (
	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/utils"
)

// GatherResources gathers log storage resources from a namespace
func GatherResources(client *oc.Client, namespace string) error {
	utils.Log("BEGIN gather_logstore_resources ...")

	utils.Log("Gathering data for logstore component")
	if err := gatherLokistack(client, namespace); err != nil {
		utils.Log("Warning: failed to gather lokistack resources: %v", err)
	}

	utils.Log("END gather_logstore_resources ...")
	return nil
}

// gatherLokistack gathers Lokistack resources
func gatherLokistack(client *oc.Client, namespace string) error {
	utils.Log("Gathering Lokistack resources")
	utils.Log("-- Gather Lokistack CR")

	adm := client.Adm()
	return adm.Inspect(oc.InspectOptions{
		Namespace: namespace,
		Resources: []string{"lokistacks.loki.grafana.com"},
	})
}
