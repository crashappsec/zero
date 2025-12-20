package report

import (
	"strings"
	"testing"
	"time"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("Expected non-nil builder")
	}
	if b.Len() != 0 {
		t.Errorf("Len() = %d, want 0", b.Len())
	}
}

func TestBuilderTitle(t *testing.T) {
	b := NewBuilder()
	b.Title("Test Title")

	result := b.String()
	if !strings.HasPrefix(result, "# Test Title\n") {
		t.Errorf("Title not formatted correctly: %s", result)
	}
}

func TestBuilderMeta(t *testing.T) {
	b := NewBuilder()
	b.Meta(ReportMeta{
		Repository:  "test/repo",
		Timestamp:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ScannerDesc: "Test Scanner",
	})

	result := b.String()

	if !strings.Contains(result, "**Repository:** `test/repo`") {
		t.Error("Expected repository in meta")
	}
	if !strings.Contains(result, "**Generated:**") {
		t.Error("Expected generated timestamp in meta")
	}
	if !strings.Contains(result, "**Scanner:** Test Scanner") {
		t.Error("Expected scanner description in meta")
	}
	if !strings.Contains(result, "---") {
		t.Error("Expected divider after meta")
	}
}

func TestBuilderSection(t *testing.T) {
	tests := []struct {
		level    int
		title    string
		expected string
	}{
		{1, "Level 1", "# Level 1\n"},
		{2, "Level 2", "## Level 2\n"},
		{3, "Level 3", "### Level 3\n"},
		{6, "Level 6", "###### Level 6\n"},
		{0, "Clamped to 1", "# Clamped to 1\n"},
		{10, "Clamped to 6", "###### Clamped to 6\n"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			b := NewBuilder()
			b.Section(tt.level, tt.title)
			result := b.String()
			if !strings.HasPrefix(result, tt.expected) {
				t.Errorf("Section(%d, %s) = %q, want prefix %q", tt.level, tt.title, result, tt.expected)
			}
		})
	}
}

func TestBuilderTable(t *testing.T) {
	b := NewBuilder()
	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"Row1", "Val1"},
		{"Row2", "Val2"},
	}
	b.Table(headers, rows)

	result := b.String()

	if !strings.Contains(result, "| Name | Value |") {
		t.Error("Expected header row")
	}
	if !strings.Contains(result, "|--------|--------|") {
		t.Error("Expected separator row")
	}
	if !strings.Contains(result, "| Row1 | Val1 |") {
		t.Error("Expected first data row")
	}
	if !strings.Contains(result, "| Row2 | Val2 |") {
		t.Error("Expected second data row")
	}
}

func TestBuilderTableEmpty(t *testing.T) {
	b := NewBuilder()
	b.Table([]string{}, nil)

	if b.Len() != 0 {
		t.Error("Expected empty table to produce no output")
	}
}

func TestBuilderTableShortRow(t *testing.T) {
	b := NewBuilder()
	headers := []string{"A", "B", "C"}
	rows := [][]string{
		{"Only", "Two"}, // Missing third column
	}
	b.Table(headers, rows)

	result := b.String()
	// Should handle missing cells gracefully
	if !strings.Contains(result, "| Only | Two |  |") {
		t.Errorf("Short row not handled correctly: %s", result)
	}
}

func TestBuilderParagraph(t *testing.T) {
	b := NewBuilder()
	b.Paragraph("This is a paragraph.")

	result := b.String()
	if result != "This is a paragraph.\n\n" {
		t.Errorf("Paragraph = %q, want 'This is a paragraph.\\n\\n'", result)
	}
}

func TestBuilderList(t *testing.T) {
	b := NewBuilder()
	b.List([]string{"Item 1", "Item 2", "Item 3"})

	result := b.String()
	if !strings.Contains(result, "- Item 1\n") {
		t.Error("Expected first list item")
	}
	if !strings.Contains(result, "- Item 2\n") {
		t.Error("Expected second list item")
	}
}

func TestBuilderNumberedList(t *testing.T) {
	b := NewBuilder()
	b.NumberedList([]string{"First", "Second", "Third"})

	result := b.String()
	if !strings.Contains(result, "1. First\n") {
		t.Error("Expected first numbered item")
	}
	if !strings.Contains(result, "2. Second\n") {
		t.Error("Expected second numbered item")
	}
	if !strings.Contains(result, "3. Third\n") {
		t.Error("Expected third numbered item")
	}
}

func TestBuilderCodeBlock(t *testing.T) {
	b := NewBuilder()
	b.CodeBlock("go", "fmt.Println(\"hello\")")

	result := b.String()
	if !strings.Contains(result, "```go\n") {
		t.Error("Expected code block opening with language")
	}
	if !strings.Contains(result, "fmt.Println(\"hello\")") {
		t.Error("Expected code content")
	}
	if !strings.Contains(result, "\n```\n") {
		t.Error("Expected code block closing")
	}
}

func TestBuilderQuote(t *testing.T) {
	b := NewBuilder()
	b.Quote("This is a quote.")

	result := b.String()
	if !strings.Contains(result, "> This is a quote.\n") {
		t.Errorf("Quote = %q", result)
	}
}

func TestBuilderQuoteMultiline(t *testing.T) {
	b := NewBuilder()
	b.Quote("Line 1\nLine 2")

	result := b.String()
	if !strings.Contains(result, "> Line 1\n") {
		t.Error("Expected first line quoted")
	}
	if !strings.Contains(result, "> Line 2\n") {
		t.Error("Expected second line quoted")
	}
}

func TestBuilderDivider(t *testing.T) {
	b := NewBuilder()
	b.Divider()

	result := b.String()
	if result != "---\n\n" {
		t.Errorf("Divider = %q, want '---\\n\\n'", result)
	}
}

func TestBuilderFormatters(t *testing.T) {
	b := NewBuilder()

	if got := b.Bold("text"); got != "**text**" {
		t.Errorf("Bold() = %s, want **text**", got)
	}

	if got := b.Italic("text"); got != "*text*" {
		t.Errorf("Italic() = %s, want *text*", got)
	}

	if got := b.Code("text"); got != "`text`" {
		t.Errorf("Code() = %s, want `text`", got)
	}

	if got := b.Link("title", "https://example.com"); got != "[title](https://example.com)" {
		t.Errorf("Link() = %s", got)
	}
}

func TestBuilderKeyValue(t *testing.T) {
	b := NewBuilder()
	b.KeyValue("Status", "Active")

	result := b.String()
	if result != "**Status:** Active\n" {
		t.Errorf("KeyValue = %q", result)
	}
}

func TestBuilderRaw(t *testing.T) {
	b := NewBuilder()
	b.Raw("raw text without formatting")

	if b.String() != "raw text without formatting" {
		t.Errorf("Raw = %q", b.String())
	}
}

func TestBuilderNewline(t *testing.T) {
	b := NewBuilder()
	b.Raw("text")
	b.Newline()

	if b.String() != "text\n" {
		t.Errorf("String = %q", b.String())
	}
}

func TestBuilderFooter(t *testing.T) {
	b := NewBuilder()
	b.Footer("SBOM")

	result := b.String()
	if !strings.Contains(result, "---\n") {
		t.Error("Expected divider in footer")
	}
	if !strings.Contains(result, "Generated by Zero SBOM Scanner") {
		t.Error("Expected scanner name in footer")
	}
}

func TestBuilderBytes(t *testing.T) {
	b := NewBuilder()
	b.Paragraph("test")

	bytes := b.Bytes()
	if string(bytes) != b.String() {
		t.Error("Bytes() should match String()")
	}
}

func TestBuilderReset(t *testing.T) {
	b := NewBuilder()
	b.Paragraph("some content")

	if b.Len() == 0 {
		t.Error("Expected content before reset")
	}

	b.Reset()

	if b.Len() != 0 {
		t.Errorf("Len() after reset = %d, want 0", b.Len())
	}
}

func TestBuilderChaining(t *testing.T) {
	b := NewBuilder()
	result := b.Title("Report").
		Section(2, "Section 1").
		Paragraph("Some text.").
		List([]string{"Item 1", "Item 2"}).
		Divider().
		Footer("Test").
		String()

	if !strings.Contains(result, "# Report") {
		t.Error("Expected title")
	}
	if !strings.Contains(result, "## Section 1") {
		t.Error("Expected section")
	}
	if !strings.Contains(result, "Some text.") {
		t.Error("Expected paragraph")
	}
	if !strings.Contains(result, "- Item 1") {
		t.Error("Expected list")
	}
}
