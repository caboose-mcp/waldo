package tui

// PageSlice returns the items for the current page and pagination state.
func PageSlice(filtered []string, start, size int) (page []string, hasPrev, hasNext bool) {
	total := len(filtered)

	if start >= total {
		start = 0
	}
	if start < 0 {
		start = 0
	}

	end := start + size
	if end > total {
		end = total
	}

	hasPrev = start > 0
	hasNext = end < total

	return filtered[start:end], hasPrev, hasNext
}

// PageCount returns the total number of pages needed to display items.
func PageCount(total, size int) int {
	if total == 0 {
		return 1
	}
	return (total + size - 1) / size
}

// CurrentPage returns the 1-indexed page number for a given start index and page size.
func CurrentPage(start, size int) int {
	return start/size + 1
}
