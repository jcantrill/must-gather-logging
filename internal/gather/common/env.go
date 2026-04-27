package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/must-gather-logging/internal/client/oc"
	"github.com/openshift/must-gather-logging/internal/utils/log"
	"github.com/openshift/must-gather-logging/internal/utils"
)

// GetEnv gets environment and build information from a pod
func GetEnv(client *oc.Client, pod, outputDir, namespace, pattern string) error {
	log.Begin(0, "get_env ...")
	envFile := filepath.Join(outputDir, pod)

	log.Log("---- Env for %s", pod)

	// Get container names
	containers, err := client.Get(oc.GetOptions{
		Resource:  "pod",
		Name:      pod,
		Namespace: namespace,
		Output:    "jsonpath={.spec.containers[*].name}",
	})

	if err != nil {
		return fmt.Errorf("failed to get containers: %w", err)
	}

	containerList := utils.ParseLines(string(containers))
	var envData strings.Builder

	for _, container := range containerList {
		if container == "" {
			continue
		}

		log.Log("----- Inspecting container %s", container)

		// Try to get build info
		dockerfile, err := client.Exec(oc.ExecOptions{
			Pod:       pod,
			Namespace: namespace,
			Container: container,
			Command:   []string{"ls", "/root/buildinfo"},
		})

		if err == nil && len(dockerfile) > 0 {
			// Filter for matching pattern
			files := utils.ParseLines(string(dockerfile))
			for _, file := range files {
				if matched, _ := filepath.Match(pattern, file); matched {
					log.Log("----- Getting buildInfo")
					envData.WriteString(fmt.Sprintf("Image info: %s\n", file))

					// Get build date
					buildDate, err := client.Exec(oc.ExecOptions{
						Pod:       pod,
						Namespace: namespace,
						Container: container,
						Command:   []string{"grep", "-o", `"build-date"="[^[:blank:]]*"`, "/root/buildinfo/" + file},
					})

					if err == nil {
						envData.Write(buildDate)
						envData.WriteString("\n")
					} else {
						log.Log("---- Unable to get build date")
					}
					break
				}
			}
		}

		// Get environment variables
		log.Log("----- Getting environment variables")
		envData.WriteString("-- Environment Variables\n")

		envVars, err := client.Exec(oc.ExecOptions{
			Pod:       pod,
			Namespace: namespace,
			Container: container,
			Command:   []string{"env"},
		})

		if err != nil {
			log.Warn("failed to get environment variables: %v", err)
		} else {
			// Sort environment variables
			lines := utils.ParseLines(string(envVars))
			sortedLines := make([]string, len(lines))
			copy(sortedLines, lines)
			// Note: In Go 1.21+, we could use slices.Sort, but for compatibility:
			for i := 0; i < len(sortedLines); i++ {
				for j := i + 1; j < len(sortedLines); j++ {
					if sortedLines[i] > sortedLines[j] {
						sortedLines[i], sortedLines[j] = sortedLines[j], sortedLines[i]
					}
				}
			}
			for _, line := range sortedLines {
				envData.WriteString(line)
				envData.WriteString("\n")
			}
		}
	}

	log.End(0, "get_env ...")

	return os.WriteFile(envFile, []byte(envData.String()), 0644)
}
