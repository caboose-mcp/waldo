package git

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type BranchStatus struct {
	Branch      string
	IsDetached  bool
	IsProd      bool
	Dirty       bool
	AheadBy     int
	BehindBy    int
	LastUpdated time.Time
}

var (
	cache      *BranchStatus
	cacheLock  sync.RWMutex
	cacheTime  time.Time
	cacheTTL   = 5 * time.Second // Rate limit: 5 second cache
)

// GetStatus returns the current git branch status with caching and rate limiting.
func GetStatus() (*BranchStatus, error) {
	cacheLock.RLock()
	if cache != nil && time.Since(cacheTime) < cacheTTL {
		defer cacheLock.RUnlock()
		return cache, nil
	}
	cacheLock.RUnlock()

	status, err := fetchStatus()
	if err != nil {
		return nil, err
	}

	cacheLock.Lock()
	cache = status
	cacheTime = time.Now()
	status.LastUpdated = cacheTime
	cacheLock.Unlock()

	return status, nil
}

func fetchStatus() (*BranchStatus, error) {
	status := &BranchStatus{}

	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	out, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return nil, fmt.Errorf("not in a git repository")
	}

	// Get current branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	branch := strings.TrimSpace(string(out))
	status.Branch = branch
	status.IsDetached = branch == "HEAD"

	// Check if branch is production (prod, main, master)
	status.IsProd = isProdBranch(branch)

	// Check if working directory is dirty
	cmd = exec.Command("git", "status", "--porcelain")
	out, err = cmd.CombinedOutput()
	if err == nil {
		status.Dirty = len(strings.TrimSpace(string(out))) > 0
	}

	// Get ahead/behind counts
	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	out, err = cmd.CombinedOutput()
	if err == nil {
		counts := strings.Fields(strings.TrimSpace(string(out)))
		if len(counts) == 2 {
			fmt.Sscanf(counts[0], "%d", &status.AheadBy)
			fmt.Sscanf(counts[1], "%d", &status.BehindBy)
		}
	}

	return status, nil
}

func isProdBranch(branch string) bool {
	prodBranches := map[string]bool{
		"main":       true,
		"master":     true,
		"prod":       true,
		"production": true,
		"release":    true,
		"stable":     true,
	}
	return prodBranches[branch]
}

// StatuslineFormat returns a formatted status string for use in status bars.
// Includes emoji warnings for prod/detached branches.
func (s *BranchStatus) StatuslineFormat() string {
	var parts []string

	if s.IsDetached {
		parts = append(parts, "🔴 DETACHED")
	} else if s.IsProd {
		parts = append(parts, fmt.Sprintf("🚨 %s", s.Branch))
	} else {
		parts = append(parts, fmt.Sprintf("🌿 %s", s.Branch))
	}

	if s.Dirty {
		parts = append(parts, "●")
	}

	if s.AheadBy > 0 {
		parts = append(parts, fmt.Sprintf("⬆ %d", s.AheadBy))
	}

	if s.BehindBy > 0 {
		parts = append(parts, fmt.Sprintf("⬇ %d", s.BehindBy))
	}

	return strings.Join(parts, " ")
}

// WarningLevel returns the visual warning level for use in terminal coloring.
// "red" for detached/prod, "yellow" for dirty, "green" for clean.
func (s *BranchStatus) WarningLevel() string {
	if s.IsDetached || s.IsProd {
		return "red"
	}
	if s.Dirty {
		return "yellow"
	}
	return "green"
}

// ClearCache invalidates the status cache, forcing a refresh on next call.
func ClearCache() {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	cache = nil
}
