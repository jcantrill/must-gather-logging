package oc

import (
	"fmt"
)

// Adm provides methods for oc adm subcommands
type Adm struct {
	client *Client
}

// Adm returns an Adm instance for executing oc adm commands
func (c *Client) Adm() *Adm {
	return &Adm{client: c}
}

// InspectOptions represents options for oc adm inspect command
type InspectOptions struct {
	// DestDir is the destination directory for collected data
	DestDir string
	// Namespace is the namespace to inspect (optional if resource is cluster-scoped)
	Namespace string
	// CacheDir is the cache directory for oc
	CacheDir string
	// Resources is the list of resources to inspect (e.g., "ns/openshift-logging", "nodes")
	Resources []string
}

// Inspect executes oc adm inspect command
func (a *Adm) Inspect(opts InspectOptions) error {
	args := []string{"adm", "inspect"}

	// Use client's cache dir if not specified in options
	cacheDir := opts.CacheDir
	if cacheDir == "" {
		cacheDir = a.client.CacheDir
	}
	if cacheDir != "" {
		args = append(args, "--cache-dir="+cacheDir)
	}

	// Use client's base path if not specified in options
	destDir := opts.DestDir
	if destDir == "" {
		destDir = a.client.BasePath
	}
	if destDir != "" {
		args = append(args, "--dest-dir="+destDir)
	}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	args = append(args, opts.Resources...)

	fullArgs := a.client.buildArgs(args)
	output, err := a.client.Execute(fullArgs...)

	// Log the output from oc adm inspect (contains "Gathering data for..." and "Wrote inspect data to..." messages)
	if len(output) > 0 {
		a.client.LogOutput(string(output))
	}

	return err
}

// MustGatherOptions represents options for oc adm must-gather command
type MustGatherOptions struct {
	// Image is the must-gather image to use
	Image string
	// ImageStream is an optional image stream to use
	ImageStream string
	// DestDir is the destination directory for collected data
	DestDir string
	// Command is the command to run in the must-gather pod (e.g., "/usr/bin/gather")
	Command string
	// NodeSelector is a label selector for node selection
	NodeSelector string
	// Timeout for the must-gather operation
	Timeout string
}

// MustGather executes oc adm must-gather command
func (a *Adm) MustGather(opts MustGatherOptions) error {
	args := []string{"adm", "must-gather"}

	if opts.Image != "" {
		args = append(args, "--image="+opts.Image)
	}

	if opts.ImageStream != "" {
		args = append(args, "--image-stream="+opts.ImageStream)
	}

	if opts.DestDir != "" {
		args = append(args, "--dest-dir="+opts.DestDir)
	}

	if opts.NodeSelector != "" {
		args = append(args, "--node-selector="+opts.NodeSelector)
	}

	if opts.Timeout != "" {
		args = append(args, "--timeout="+opts.Timeout)
	}

	if opts.Command != "" {
		args = append(args, "--", opts.Command)
	}

	fullArgs := a.client.buildArgs(args)
	return a.client.ExecuteWithStdout(fullArgs...)
}

// NodeLogsOptions represents options for oc adm node-logs command
type NodeLogsOptions struct {
	// NodeName is the name of the node
	NodeName string
	// Unit is the systemd unit to get logs from
	Unit string
	// Path is the path to the log file
	Path string
	// Tail is the number of lines to show from the end
	Tail int
}

// NodeLogs executes oc adm node-logs command
func (a *Adm) NodeLogs(opts NodeLogsOptions) ([]byte, error) {
	if opts.NodeName == "" {
		return nil, fmt.Errorf("node name is required")
	}

	args := []string{"adm", "node-logs", opts.NodeName}

	if opts.Unit != "" {
		args = append(args, "-u", opts.Unit)
	}

	if opts.Path != "" {
		args = append(args, "--path="+opts.Path)
	}

	if opts.Tail > 0 {
		args = append(args, fmt.Sprintf("--tail=%d", opts.Tail))
	}

	fullArgs := a.client.buildArgs(args)
	return a.client.Execute(fullArgs...)
}

// TopOptions represents options for oc adm top command
type TopOptions struct {
	// Resource is the resource type (node or pod)
	Resource string
	// Namespace for pod metrics
	Namespace string
	// Selector for filtering
	Selector string
	// ShowContainers shows container-level metrics
	ShowContainers bool
}

// Top executes oc adm top command
func (a *Adm) Top(opts TopOptions) ([]byte, error) {
	if opts.Resource == "" {
		return nil, fmt.Errorf("resource type is required (node or pod)")
	}

	args := []string{"adm", "top", opts.Resource}

	if opts.Namespace != "" {
		args = append(args, "-n", opts.Namespace)
	}

	if opts.Selector != "" {
		args = append(args, "--selector="+opts.Selector)
	}

	if opts.ShowContainers {
		args = append(args, "--containers")
	}

	fullArgs := a.client.buildArgs(args)
	return a.client.Execute(fullArgs...)
}
