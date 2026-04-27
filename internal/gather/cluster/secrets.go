package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/openshift/must-gather-logging/internal/utils"
)

// redactSecretsInNamespaces finds and redacts all secret files in the collected namespaces
func redactSecretsInNamespaces(basePath string, namespaces []string) error {
	utils.Log("BEGIN redacting secrets ...")

	var wg sync.WaitGroup
	for _, ns := range namespaces {
		wg.Add(1)
		go func(namespace string) {
			defer wg.Done()

			// Redact aggregated secrets.yaml file
			secretsFile := filepath.Join(basePath, "namespaces", namespace, "core", "secrets.yaml")
			if err := redactAggregatedSecretsFile(secretsFile); err != nil {
				utils.Log("Warning: failed to redact secrets.yaml in namespace %s: %v", namespace, err)
			}

			// Redact individual secret files if they exist
			secretsDir := filepath.Join(basePath, "namespaces", namespace, "core", "secrets")
			if err := redactSecretsInDirectory(secretsDir); err != nil {
				utils.Log("Warning: failed to redact secrets directory in namespace %s: %v", namespace, err)
			}
		}(ns)
	}

	wg.Wait()
	utils.Log("END redacting secrets ...")
	return nil
}

// redactAggregatedSecretsFile redacts sensitive data in an aggregated secrets.yaml file
func redactAggregatedSecretsFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to redact
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	var secretList map[string]interface{}
	if err := yaml.Unmarshal(data, &secretList); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Check if this is a List resource (List or SecretList)
	kind, ok := secretList["kind"].(string)
	if !ok || (kind != "List" && kind != "SecretList") {
		// Might be a single secret, try to redact as single secret
		return redactSecretInData(&secretList)
	}

	// Process items list
	items, ok := secretList["items"].([]interface{})
	if !ok {
		return nil // No items, nothing to redact
	}

	// Redact each secret in the list
	for i := range items {
		if secretMap, ok := items[i].(map[string]interface{}); ok {
			if err := redactSecretInData(&secretMap); err != nil {
				utils.Log("Warning: failed to redact secret in list: %v", err)
			}
		}
	}

	// Marshal back to YAML
	redactedData, err := yaml.Marshal(secretList)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, redactedData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// redactSecretInData redacts a single secret's data fields
func redactSecretInData(secret *map[string]interface{}) error {
	// Check if this is a Secret resource
	kind, ok := (*secret)["kind"].(string)
	if !ok || kind != "Secret" {
		return nil // Not a secret, skip
	}

	// Redact data field
	if dataField, ok := (*secret)["data"].(map[string]interface{}); ok {
		for key := range dataField {
			dataField[key] = "REDACTED"
		}
	}

	// Redact stringData field
	if stringDataField, ok := (*secret)["stringData"].(map[string]interface{}); ok {
		for key := range stringDataField {
			stringDataField[key] = "REDACTED"
		}
	}

	return nil
}

// redactSecretsInDirectory redacts all secret YAML files in a directory
func redactSecretsInDirectory(dirPath string) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to redact
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read secrets directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		secretFile := filepath.Join(dirPath, entry.Name())
		if err := redactSecretFile(secretFile); err != nil {
			utils.Log("Warning: failed to redact secret file %s: %v", secretFile, err)
		}
	}

	return nil
}

// redactSecretFile redacts sensitive data in a secret YAML file
func redactSecretFile(filePath string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	var secret map[string]interface{}
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Check if this is a Secret resource
	kind, ok := secret["kind"].(string)
	if !ok || kind != "Secret" {
		return nil // Not a secret, skip
	}

	// Redact data field
	if dataField, ok := secret["data"].(map[string]interface{}); ok {
		for key := range dataField {
			dataField[key] = "REDACTED"
		}
	}

	// Redact stringData field
	if stringDataField, ok := secret["stringData"].(map[string]interface{}); ok {
		for key := range stringDataField {
			stringDataField[key] = "REDACTED"
		}
	}

	// Marshal back to YAML
	redactedData, err := yaml.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, redactedData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
