package storage

import (
	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	"github.com/openshift/must-gather-logging/internal/utils/log"
)

const (
	KindLokiStack = "lokistacks.loki.grafana.com"
)

// GatherResources gathers log storage resources from a namespace
func GatherResources(client *oc.Client, namespace string) error {
	exists, err := common.HasCrd(client, KindLokiStack)
	if !exists {
		return err
	}
	log.Begin(0, "gather_logstore_resources ...")

	log.Log("Gathering data for logstore component")
	if err := gatherLokistack(client, namespace); err != nil {
		log.Warn("failed to gather lokistack resources: %v", err)
	}

	log.End(0, "gather_logstore_resources ...")
	return nil
}

// gatherLokistack gathers Lokistack resources
func gatherLokistack(client *oc.Client, namespace string) error {
	log.Log("Gathering Lokistack resources")
	log.Log("-- Gather Lokistack CR")

	adm := client.Adm()
	return adm.Inspect(oc.InspectOptions{
		Namespace: namespace,
		Resources: []string{"lokistacks.loki.grafana.com"},
	})
}
