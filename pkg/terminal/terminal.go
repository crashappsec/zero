// Package terminal provides colored output and progress display
package terminal

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Red       = "\033[0;31m"
	Green     = "\033[0;32m"
	Yellow    = "\033[1;33m"
	Blue      = "\033[0;34m"
	Cyan      = "\033[0;36m"
	White     = "\033[0;37m"
	BoldRed   = "\033[1;31m"
	BoldGreen = "\033[1;32m"
)

// Icons for status display
const (
	IconSuccess  = "✓"
	IconFailed   = "✗"
	IconRunning  = "◐"
	IconQueued   = "○"
	IconSkipped  = "⊘"
	IconWarning  = "⚠"
	IconArrow    = "▸"
)

// Terminal provides thread-safe terminal output
type Terminal struct {
	mu       sync.Mutex
	noColor  bool
	width    int
}

// New creates a new Terminal instance
func New() *Terminal {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	return &Terminal{
		noColor: os.Getenv("NO_COLOR") != "",
		width:   width,
	}
}

// Color wraps text in color codes if colors are enabled
func (t *Terminal) Color(code, text string) string {
	if t.noColor {
		return text
	}
	return code + text + Reset
}

// Success prints a success message
func (t *Terminal) Success(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", t.Color(Green, IconSuccess), msg)
}

// Error prints an error message
func (t *Terminal) Error(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", t.Color(Red, IconFailed), msg)
}

// Warning prints a warning message
func (t *Terminal) Warning(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s  %s\n", t.Color(Yellow, IconWarning), t.Color(Bold, msg))
}

// Info prints an info message
func (t *Terminal) Info(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf(format+"\n", args...)
}

// Header prints a section header
func (t *Terminal) Header(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("\n%s\n\n", t.Color(Bold, text))
}

// SubHeader prints a sub-section header
func (t *Terminal) SubHeader(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("%s %s\n", t.Color(Bold, text), t.Color(Dim, "(depth=1)"))
}

// Divider prints a horizontal line
func (t *Terminal) Divider() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Println(strings.Repeat("━", min(t.width, 78)))
}

// RepoCloned prints a cloned repo result
func (t *Terminal) RepoCloned(name, size, files, commit, status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("  %s %s %s %s\n",
		t.Color(Green, IconSuccess),
		name,
		t.Color(Dim, fmt.Sprintf("%s, %s files", size, files)),
		t.Color(Dim, fmt.Sprintf("(%s %s)", commit, status)),
	)
}

// RepoScanning prints a repo that's being scanned
func (t *Terminal) RepoScanning(name string, estimate int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if estimate > 10 {
		fmt.Printf("  %s %s %s\n",
			t.Color(Yellow, IconRunning),
			t.Color(Bold, name),
			t.Color(Dim, fmt.Sprintf("scanning (~%ds)...", estimate)),
		)
	} else {
		fmt.Printf("  %s %s %s\n",
			t.Color(Yellow, IconRunning),
			t.Color(Bold, name),
			t.Color(Dim, "scanning..."),
		)
	}
}

// ScannerRunning prints a running scanner line
func (t *Terminal) ScannerRunning(name string, estimate int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s %s\n",
		t.Color(Cyan, IconArrow),
		name,
		t.Color(Cyan, "running"),
		t.Color(Dim, fmt.Sprintf("~%ds", estimate)),
	)
}

// ScannerQueued prints a queued scanner line
func (t *Terminal) ScannerQueued(name string, estimate int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s\n",
		t.Color(Dim, IconQueued),
		t.Color(Dim, name),
		t.Color(Dim, fmt.Sprintf("queued  ~%ds", estimate)),
	)
}

// ScannerSkipped prints a skipped scanner line
func (t *Terminal) ScannerSkipped(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s\n",
		t.Color(Dim, IconSkipped),
		t.Color(Dim, name),
		t.Color(Dim, "skipped"),
	)
}

// ScannerComplete prints a completed scanner result
func (t *Terminal) ScannerComplete(name, summary string, duration int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("      %s %-20s %s %s\n",
		t.Color(Green, IconSuccess),
		name,
		summary,
		t.Color(Dim, fmt.Sprintf("%ds", duration)),
	)
}

// Progress prints an in-place progress line (overwrites current line)
func (t *Terminal) Progress(completed, total int, active string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("\r\033[K  %s",
		t.Color(Dim, fmt.Sprintf("[%d/%d scanners] %s", completed, total, active)),
	)
}

// ClearLine clears the current line
func (t *Terminal) ClearLine() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Print("\r\033[K")
}

// RepoComplete prints a completed repo header
func (t *Terminal) RepoComplete(name string, success bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if success {
		fmt.Printf("  %s %s\n", t.Color(Green, IconSuccess), t.Color(Bold, name))
	} else {
		fmt.Printf("  %s %s %s\n", t.Color(Red, IconFailed), t.Color(Bold, name), t.Color(Red, "(scan failed)"))
	}
}

// ScanComplete prints the final scanning complete message
func (t *Terminal) ScanComplete() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Printf("%s %s\n", t.Color(Green, IconSuccess), t.Color(Bold, "Scanning complete"))
}

// Summary prints the hydrate summary
func (t *Terminal) Summary(org string, duration int, success, failed int, diskUsage, files string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Printf("\n%s\n\n", t.Color(Bold+Green, "✓ Hydrate Complete"))

	fmt.Printf("%s\n", t.Color(Bold, "Summary"))
	fmt.Printf("  Organization:    %s\n", t.Color(Cyan, org))
	fmt.Printf("  Duration:        %ds\n", duration)
	if failed > 0 {
		fmt.Printf("  Repos scanned:   %s, %s\n",
			t.Color(Green, fmt.Sprintf("%d success", success)),
			t.Color(Red, fmt.Sprintf("%d failed", failed)),
		)
	} else {
		fmt.Printf("  Repos scanned:   %s\n", t.Color(Green, fmt.Sprintf("%d success", success)))
	}
	fmt.Printf("  Disk usage:      %s\n", diskUsage)
	fmt.Printf("  Total files:     %s\n", files)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
