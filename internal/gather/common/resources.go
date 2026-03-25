package common

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/utils"
)

const (
	ColumnsMetadataName      = "custom-columns=:.metadata.name"
	ColumnsMetadataNamespace = "custom-columns=:.metadata.namespace"
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

// GetResourceNames returns resource names using custom-columns output
func GetResourceNames(client *oc.Client, resource, namespace string) ([]string, error) {
	opts := oc.GetOptions{
		Resource:       resource,
		Output:         ColumnsMetadataName,
		NoHeaders:      true,
		IgnoreNotFound: true,
	}

	if namespace == "" {
		opts.AllNamespaces = true
	} else {
		opts.Namespace = namespace
	}

	result, err := client.Get(opts)
	if err != nil {
		return nil, err
	}

	return utils.ParseLines(string(result)), nil
}

// GetResourceNamespaces returns namespaces where a resource exists using custom-columns output
func GetResourceNamespaces(client *oc.Client, resource string) (mapset.Set[string], error) {
	result, err := client.Get(oc.GetOptions{
		Resource:       resource,
		AllNamespaces:  true,
		Output:         ColumnsMetadataNamespace,
		NoHeaders:      true,
		IgnoreNotFound: true,
	})

	if err != nil {
		return nil, err
	}
	lines := utils.ParseLines(string(result))
	return mapset.NewSet(lines...), nil
}
