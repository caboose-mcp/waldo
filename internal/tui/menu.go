package tui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Menu struct {
	Items     []string
	Filtered  []string
	Cursor    int
	PageStart int
	PageSize  int
	Query     string
	Title     string
}

func NewMenu(items []string, title string, pageSize int) *Menu {
	return &Menu{
		Items:    items,
		Filtered: items,
		Title:    title,
		PageSize: pageSize,
		Cursor:   0,
		Query:    "",
	}
}

func (m *Menu) Run(stdin io.Reader) (selected string, cancelled bool, err error) {
	file, ok := stdin.(*os.File)
	if !ok {
		return "", false, fmt.Errorf("stdin must be a file")
	}

	oldState, err := term.MakeRaw(int(file.Fd()))
	if err != nil {
		return "", false, err
	}
	defer term.Restore(int(file.Fd()), oldState)

	m.render()
	buf := make([]byte, 8)

	for {
		n, err := file.Read(buf)
		if err != nil {
			return "", false, err
		}

		if n == 0 {
			continue
		}

		ev, r := parseKey(buf[:n])

		switch ev {
		case KeyEsc:
			m.clearRender()
			return "", true, nil

		case KeyEnter:
			if len(m.Filtered) > 0 {
				selected = m.Filtered[m.PageStart+m.Cursor]
				m.clearRender()
				return selected, false, nil
			}

		case KeyUp:
			m.moveCursor(-1)

		case KeyDown:
			m.moveCursor(1)

		case KeyBackspace:
			if len(m.Query) > 0 {
				m.Query = m.Query[:len(m.Query)-1]
				m.refilter()
			}

		case KeyRune:
			m.Query += string(r)
			m.refilter()
		}

		m.render()
	}
}

func (m *Menu) moveCursor(delta int) {
	pageItems, _, hasNext := PageSlice(m.Filtered, m.PageStart, m.PageSize)
	newCursor := m.Cursor + delta

	if newCursor >= len(pageItems) {
		if hasNext {
			m.PageStart += m.PageSize
			if m.PageStart >= len(m.Filtered) {
				m.PageStart = len(m.Filtered) - m.PageSize
			}
			m.Cursor = 0
		}
		return
	}

	if newCursor < 0 {
		if m.PageStart > 0 {
			m.PageStart -= m.PageSize
			if m.PageStart < 0 {
				m.PageStart = 0
			}
			m.Cursor = m.PageSize - 1
		}
		return
	}

	m.Cursor = newCursor
}

func (m *Menu) refilter() {
	m.Filtered = FuzzyFilter(m.Items, m.Query)
	m.PageStart = 0
	m.Cursor = 0
}

func (m *Menu) render() {
	lines := m.drawLines()
	for _, line := range lines {
		fmt.Println(line)
	}
}

func (m *Menu) clearRender() {
	pageItems, _, _ := PageSlice(m.Filtered, m.PageStart, m.PageSize)
	numLines := len(pageItems) + 3 // title + filter + footer
	for i := 0; i < numLines; i++ {
		fmt.Print("\033[1A\033[2K")
	}
}

func (m *Menu) drawLines() []string {
	var lines []string

	// Title
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(m.Title))

	// Filter
	filterLine := fmt.Sprintf("Filter: %s", m.Query)
	if m.Query == "" {
		filterLine = "Filter: (type to search)"
	}
	lines = append(lines, lipgloss.NewStyle().Faint(true).Render(filterLine))

	// Items (current page)
	pageItems, _, _ := PageSlice(m.Filtered, m.PageStart, m.PageSize)
	for i, item := range pageItems {
		marker := "  "
		if i == m.Cursor {
			marker = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s", marker, item))
	}

	// Footer
	total := len(m.Filtered)
	if total == 0 {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render("(no matches)"))
	} else {
		pageNum := CurrentPage(m.PageStart, m.PageSize)
		pageCount := PageCount(total, m.PageSize)
		footer := fmt.Sprintf("Page %d/%d (↑↓ navigate, type to filter, Enter select, Esc cancel)", pageNum, pageCount)
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(footer))
	}

	return lines
}

type KeyEvent int

const (
	KeyUp KeyEvent = iota
	KeyDown
	KeyEnter
	KeyEsc
	KeyBackspace
	KeyRune
	KeyOther
)

func parseKey(buf []byte) (KeyEvent, rune) {
	if len(buf) == 0 {
		return KeyOther, 0
	}

	// Single byte — printable or control
	if len(buf) == 1 {
		b := buf[0]
		if b == 27 { // ESC
			return KeyEsc, 0
		}
		if b == 13 { // Enter (CR)
			return KeyEnter, 0
		}
		if b == 127 || b == 8 { // Backspace (DEL or ^H)
			return KeyBackspace, 0
		}
		if b >= 32 && b < 127 {
			return KeyRune, rune(b)
		}
		return KeyOther, 0
	}

	// Arrow keys — ESC [ A/B/C/D
	if len(buf) >= 3 && buf[0] == 27 && buf[1] == 91 {
		switch buf[2] {
		case 65:
			return KeyUp, 0
		case 66:
			return KeyDown, 0
		}
	}

	// ESC alone (timeout case)
	if buf[0] == 27 {
		return KeyEsc, 0
	}

	return KeyOther, 0
}
