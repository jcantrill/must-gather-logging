package oc

import (
	"context"
	"fmt"
	"os/exec"
)

// Client represents an oc CLI client
type Client struct {
	// BinaryPath is the path to the oc binary (default: "oc")
	BinaryPath string
	// Namespace is the default namespace for commands
	Namespace string
	// Context is the kubeconfig context to use
	Context string
}

// NewClient creates a new oc client with default settings
func NewClient() *Client {
	return &Client{
		BinaryPath: "oc",
	}
}

// WithNamespace returns a new client with the specified namespace
func (c *Client) WithNamespace(namespace string) *Client {
	newClient := *c
	newClient.Namespace = namespace
	return &newClient
}

// WithContext returns a new client with the specified kubeconfig context
func (c *Client) WithContext(ctx string) *Client {
	newClient := *c
	newClient.Context = ctx
	return &newClient
}

// Execute runs an oc command and returns the output
func (c *Client) Execute(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("oc command failed: %w\nOutput: %s", err, string(output))
	}
	return output, nil
}

// ExecuteWithStdout runs an oc command and streams output to stdout/stderr
func (c *Client) ExecuteWithStdout(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
	cmd.Stdout = nil // Will use parent's stdout
	cmd.Stderr = nil // Will use parent's stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oc command failed: %w", err)
	}
	return nil
}

// buildArgs prepares common arguments for oc commands
func (c *Client) buildArgs(args []string) []string {
	var fullArgs []string

	if c.Context != "" {
		fullArgs = append(fullArgs, "--context", c.Context)
	}

	fullArgs = append(fullArgs, args...)

	return fullArgs
}
