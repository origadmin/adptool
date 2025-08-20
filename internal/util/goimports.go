package util

import (
	"fmt"
	"os/exec"
)

// RunGoImports formats the Go file at the given path using goimports.
// It returns an error if goimports fails or is not found.
func RunGoImports(filePath string) error {
	goimportsCmd := "goimports"
	_, err := exec.LookPath(goimportsCmd)
	if err != nil {
		// goimports not found in PATH, try go run
		goimportsCmd = "go run golang.org/x/tools/cmd/goimports"
	}

	cmd := exec.Command(goimportsCmd, "-w", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("goimports failed for %s: %s\n%s", filePath, err, string(output))
	}
	return nil
}
