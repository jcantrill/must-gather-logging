package oc

import (
	"fmt"
)

// GetOptions represents options for oc get command
type GetOptions struct {
	// Resource is the resource type (e.g., "pods", "nodes", "clusterlogforwarder")
	Resource string
	// Name is the specific resource name (optional)
	Name string
	// Namespace is the namespace (if applicable)
	Namespace string
	// AllNamespaces queries all namespaces
	AllNamespaces bool
	// Output format (yaml, json, wide, custom-columns, etc.)
	Output string
	// Selector is a label selector
	Selector string
	// FieldSelector is a field selector
	FieldSelector string
	// IgnoreNotFound ignores not found errors
	IgnoreNotFound bool
	// NoHeaders suppresses headers in output
	NoHeaders bool
}

// Get executes oc get command
func (c *Client) Get(opts GetOptions) ([]byte, error) {
	if opts.Resource == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	args := []string{"get", opts.Resource}

	if opts.Name != "" {
		args = append(args, opts.Name)
	}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	} else if opts.AllNamespaces {
		args = append(args, "-A")
	}

	if opts.Output != "" {
		args = append(args, "-o", opts.Output)
	}

	if opts.Selector != "" {
		args = append(args, "--selector="+opts.Selector)
	}

	if opts.FieldSelector != "" {
		args = append(args, "--field-selector="+opts.FieldSelector)
	}

	if opts.IgnoreNotFound {
		args = append(args, "--ignore-not-found")
	}

	if opts.NoHeaders {
		args = append(args, "--no-headers")
	}

	fullArgs := c.buildArgs(args)
	return c.Execute(fullArgs...)
}

// DescribeOptions represents options for oc describe command
type DescribeOptions struct {
	// Resource is the resource type
	Resource string
	// Name is the resource name
	Name string
	// Namespace is the namespace
	Namespace string
}

// Describe executes oc describe command
func (c *Client) Describe(opts DescribeOptions) ([]byte, error) {
	if opts.Resource == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	args := []string{"describe", opts.Resource}

	if opts.Name != "" {
		args = append(args, opts.Name)
	}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	fullArgs := c.buildArgs(args)
	return c.Execute(fullArgs...)
}

// ExecOptions represents options for oc exec command
type ExecOptions struct {
	// Pod is the pod name
	Pod string
	// Namespace is the namespace
	Namespace string
	// Container is the container name (optional)
	Container string
	// Command is the command to execute
	Command []string
}

// Exec executes oc exec command
func (c *Client) Exec(opts ExecOptions) ([]byte, error) {
	if opts.Pod == "" {
		return nil, fmt.Errorf("pod name is required")
	}

	if len(opts.Command) == 0 {
		return nil, fmt.Errorf("command is required")
	}

	args := []string{"exec"}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	if opts.Container != "" {
		args = append(args, "-c", opts.Container)
	}

	args = append(args, opts.Pod, "--")
	args = append(args, opts.Command...)

	fullArgs := c.buildArgs(args)
	return c.Execute(fullArgs...)
}
