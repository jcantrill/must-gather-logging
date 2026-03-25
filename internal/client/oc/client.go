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
	// KubeContext is the kubeconfig context to use
	KubeContext string
	// BasePath is the base collection path for must-gather operations
	BasePath string
	// CacheDir is the cache directory for oc commands
	CacheDir string
	// ctx is the context for command execution
	ctx context.Context
}

// NewClient creates a new oc client with context, base path, and cache directory
func NewClient(ctx context.Context, basePath, cacheDir string) *Client {
	return &Client{
		BinaryPath: "oc",
		BasePath:   basePath,
		CacheDir:   cacheDir,
		ctx:        ctx,
	}
}

// WithNamespace returns a new client with the specified namespace
func (c *Client) WithNamespace(namespace string) *Client {
	newClient := *c
	newClient.Namespace = namespace
	return &newClient
}

// Execute runs an oc command and returns the output
func (c *Client) Execute(args ...string) ([]byte, error) {
	cmd := exec.CommandContext(c.ctx, c.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("oc command failed: %w\nOutput: %s", err, string(output))
	}
	return output, nil
}

// ExecuteWithStdout runs an oc command and streams output to stdout/stderr
func (c *Client) ExecuteWithStdout(args ...string) error {
	cmd := exec.CommandContext(c.ctx, c.BinaryPath, args...)
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

	if c.KubeContext != "" {
		fullArgs = append(fullArgs, "--context", c.KubeContext)
	}

	fullArgs = append(fullArgs, args...)

	return fullArgs
}
