package common

import (
	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/utils"
)

const (
	ColumnsMetadataName = "custom-columns=:.metadata.name"
)

// GetPodsBySelector returns pod names matching the given selector in a namespace
func GetPodsBySelector(client *oc.Client, namespace, selector string) ([]string, error) {
	pods, err := client.Get(oc.GetOptions{
		Resource:       "pods",
		Namespace:      namespace,
		Selector:       selector,
		Output:         ColumnsMetadataName,
		NoHeaders:      true,
		IgnoreNotFound: true,
	})

	if err != nil {
		return nil, err
	}

	return utils.ParseLines(string(pods)), nil
}

// GetPodsBySelectorWithFieldSelector returns pod names matching the given selector and field selector
func GetPodsBySelectorWithFieldSelector(client *oc.Client, namespace, selector, fieldSelector string) ([]string, error) {
	pods, err := client.Get(oc.GetOptions{
		Resource:       "pods",
		Namespace:      namespace,
		Selector:       selector,
		FieldSelector:  fieldSelector,
		Output:         ColumnsMetadataName,
		NoHeaders:      true,
		IgnoreNotFound: true,
	})

	if err != nil {
		return nil, err
	}

	return utils.ParseLines(string(pods)), nil
}
