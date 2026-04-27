package console

import (
	"fmt"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/gather/common"
	"github.com/openshift/must-gather-logging/internal/utils/log"
)

const (
	KindConsolePlugin = "consoleplugins.console.openshift.io"
	PluginName        = "logging-view-plugin"
)

func isPluginDeployed(client *oc.Client) (bool, error) {
	namespaces, err := common.GetResourceNamespaces(client, KindConsolePlugin)
	if err != nil {
		return false, err
	}
	for _, ns := range namespaces.ToSlice() {
		names, err := common.GetResourceNames(client, KindConsolePlugin, ns)
		if err != nil {
			return false, err
		}
		for _, name := range names {
			if name == PluginName {
				return true, nil
			}
		}
	}
	return false, nil
}

// GatherUIPlugin gathers UIPlugin and console resources if present
func GatherUIPlugin(client *oc.Client) error {
	exists, err := common.HasCrd(client, KindConsolePlugin)
	if !exists {
		return err
	}
	isPluginDeployed, err := isPluginDeployed(client)
	if err != nil {
		return err
	}
	if !isPluginDeployed {
		return nil
	}

	log.Begin(0, "gathering uiplugin and console resources ...")

	adm := client.Adm()

	// Inspect UIPlugin
	if err := adm.Inspect(oc.InspectOptions{
		Resources: []string{fmt.Sprintf("%s/%s", KindConsolePlugin, PluginName)},
	}); err != nil {
		log.Warn("failed to inspect uiplugin: %v", err)
	}

	// Inspect console cluster operator
	if err := adm.Inspect(oc.InspectOptions{
		Resources: []string{"co/console"},
	}); err != nil {
		log.Warn("failed to inspect console: %v", err)
	}

	log.End(0, "gathering uiplugin and console resources ...")
	return nil
}
