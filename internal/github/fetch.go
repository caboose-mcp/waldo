package github

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Script struct {
	URL     string
	Content []byte
}

// Fetch downloads a script from a GitHub URL or "owner/repo/path/to/file.sh" shorthand.
// Shorthand is expanded to: https://raw.githubusercontent.com/owner/repo/main/path/to/file.sh
func Fetch(rawURL string) (*Script, error) {
	url := normalizeURL(rawURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return &Script{URL: url, Content: content}, nil
}

func normalizeURL(s string) string {
	if strings.HasPrefix(s, "https://") {
		return s
	}

	// Parse "owner/repo/path/to/file.sh"
	parts := strings.SplitN(s, "/", 3)
	if len(parts) == 3 {
		return fmt.Sprintf(
			"https://raw.githubusercontent.com/%s/%s/main/%s",
			parts[0], parts[1], parts[2])
	}

	return s
}

// Lines returns the script split by newline.
func (s *Script) Lines() []string {
	return strings.Split(string(s.Content), "\n")
}

// Preview prints the script with line numbers.
func Preview(s *Script, w io.Writer) {
	lines := s.Lines()
	for i, line := range lines {
		fmt.Fprintf(w, "%4d  %s\n", i+1, line)
	}
}

// Confirm asks the user y/n without raw mode.
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	resp, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(resp)) == "y"
}

// Execute writes the script to a temp file, makes it executable, and runs it.
// Inherits stdin/stdout/stderr from the parent process.
// Temp file is cleaned up after execution.
func (s *Script) Execute(args []string) error {
	tmp, err := os.CreateTemp("", "waldo-script-*.sh")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(s.Content); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	// Make executable: owner read/write/execute
	if err := os.Chmod(tmpPath, 0700); err != nil {
		return err
	}

	cmd := exec.Command(tmpPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
