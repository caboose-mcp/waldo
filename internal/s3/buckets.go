package s3

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Bucket struct {
	Name         string    `json:"Name"`
	CreationDate time.Time `json:"CreationDate"`
}

// List fetches S3 buckets from AWS account.
// Respects AWS_PROFILE env var if set.
func List() ([]Bucket, error) {
	// Check if aws CLI is available
	if _, err := exec.LookPath("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found; install from https://aws.amazon.com/cli/")
	}

	args := []string{"s3api", "list-buckets", "--output", "json"}

	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		args = append([]string{"--profile", profile}, args...)
	}

	cmd := exec.Command("aws", args...)
	out, err := cmd.Output()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("aws CLI error: %s", exitErr.Stderr)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var resp struct {
		Buckets []Bucket `json:"Buckets"`
	}

	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return resp.Buckets, nil
}

// Names returns a flat slice of bucket names for use in TUI.
func Names(buckets []Bucket) []string {
	names := make([]string, len(buckets))
	for i, b := range buckets {
		names[i] = b.Name
	}
	return names
}
