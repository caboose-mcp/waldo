package tui

import (
	"sort"
	"strings"
)

type Match struct {
	Item  string
	Score int
}

// FuzzyFilter returns items matching the query, sorted by relevance.
// Prefix matches and contiguous runs are scored higher.
func FuzzyFilter(items []string, query string) []string {
	if query == "" {
		return items
	}

	q := strings.ToLower(query)
	var matches []Match

	for _, item := range items {
		lower := strings.ToLower(item)
		if !strings.Contains(lower, q) {
			continue
		}

		score := 0

		// Prefix match bonus
		if strings.HasPrefix(lower, q) {
			score += 100
		}

		// Position bonus: earlier matches rank higher
		idx := strings.Index(lower, q)
		score += len(q)*10 - idx

		matches = append(matches, Match{item, score})
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	result := make([]string, len(matches))
	for i, m := range matches {
		result[i] = m.Item
	}

	return result
}
