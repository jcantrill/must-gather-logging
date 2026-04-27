package common

import (
	"fmt"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/utils/log"
	"github.com/openshift/must-gather-logging/internal/utils"
)

func HasCrd(client *oc.Client, name string) (bool, error) {
	out, err := client.Get(oc.GetOptions{
		Resource:       name,
		Output:         ColumnsMetadataName,
		NoHeaders:      true,
		IgnoreNotFound: true,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get crd %q: %w", name, err)
	}
	crds := utils.ParseLines(string(out))
	if len(crds) == 0 {
		log.Log("No crd %q found", name)
		return false, nil
	}
	return true, nil
}

const DefaultNamespace = "openshift-logging"
