package findings

import (
	"testing"
)

func TestComputeConfidence(t *testing.T) {
	tests := []struct {
		name          string
		signals       []Signal
		expectedScore int
		expectedLevel string
	}{
		{
			name:          "no signals",
			signals:       []Signal{},
			expectedScore: 50,
			expectedLevel: "medium",
		},
		{
			name: "single high weight signal",
			signals: []Signal{
				{Type: "package", Weight: 0.9},
			},
			expectedScore: 90,
			expectedLevel: "high",
		},
		{
			name: "multiple signals",
			signals: []Signal{
				{Type: "package", Weight: 0.4},
				{Type: "import", Weight: 0.25},
				{Type: "config", Weight: 0.25},
			},
			expectedScore: 90,
			expectedLevel: "high",
		},
		{
			name: "low confidence signals",
			signals: []Signal{
				{Type: "file_extension", Weight: 0.2},
				{Type: "keyword", Weight: 0.1},
			},
			expectedScore: 30,
			expectedLevel: "low",
		},
		{
			name: "capped at 100",
			signals: []Signal{
				{Type: "package", Weight: 0.5},
				{Type: "import", Weight: 0.4},
				{Type: "config", Weight: 0.3},
			},
			expectedScore: 100,
			expectedLevel: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeConfidence(tt.signals)

			if result.Score != tt.expectedScore {
				t.Errorf("Score = %d, want %d", result.Score, tt.expectedScore)
			}
			if result.Level != tt.expectedLevel {
				t.Errorf("Level = %s, want %s", result.Level, tt.expectedLevel)
			}
			if len(result.Signals) != len(tt.signals) {
				t.Errorf("Signals count = %d, want %d", len(result.Signals), len(tt.signals))
			}
		})
	}
}

func TestDetectContext(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		lineContent string
		expectTest  bool
		expectDocs  bool
		expectEx    bool
		expectComm  bool
	}{
		{
			name:       "go test file",
			filePath:   "pkg/scanner/runner_test.go",
			expectTest: true,
		},
		{
			name:       "js test file",
			filePath:   "src/components/Button.test.tsx",
			expectTest: true,
		},
		{
			name:       "jest test dir",
			filePath:   "src/__tests__/utils.js",
			expectTest: true,
		},
		{
			name:       "spec file",
			filePath:   "spec/models/user.spec.rb",
			expectTest: true,
		},
		{
			name:       "test directory",
			filePath:   "project/tests/unit/test_parser.py",
			expectTest: true,
		},
		{
			name:       "markdown docs",
			filePath:   "project/docs/README.md",
			expectDocs: true,
		},
		{
			name:       "readme file",
			filePath:   "README.md",
			expectDocs: true,
		},
		{
			name:       "example dir",
			filePath:   "project/examples/basic/main.go",
			expectEx:   true,
		},
		{
			name:       "sample dir",
			filePath:   "project/samples/demo/app.py",
			expectEx:   true,
		},
		{
			name:        "go comment",
			filePath:    "main.go",
			lineContent: "// TODO: fix this hardcoded key",
			expectComm:  true,
		},
		{
			name:        "python comment",
			filePath:    "script.py",
			lineContent: "# API_KEY = 'secret123'",
			expectComm:  true,
		},
		{
			name:        "c-style comment",
			filePath:    "app.c",
			lineContent: "/* password = test */",
			expectComm:  true,
		},
		{
			name:        "jsdoc comment",
			filePath:    "lib.js",
			lineContent: " * @param key - the API key",
			expectComm:  true,
		},
		{
			name:       "regular source file",
			filePath:   "src/main.go",
			expectTest: false,
			expectDocs: false,
			expectEx:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := DetectContext(tt.filePath, tt.lineContent)

			if ctx.InTest != tt.expectTest {
				t.Errorf("InTest = %v, want %v", ctx.InTest, tt.expectTest)
			}
			if ctx.InDocs != tt.expectDocs {
				t.Errorf("InDocs = %v, want %v", ctx.InDocs, tt.expectDocs)
			}
			if ctx.InExample != tt.expectEx {
				t.Errorf("InExample = %v, want %v", ctx.InExample, tt.expectEx)
			}
			if ctx.InComment != tt.expectComm {
				t.Errorf("InComment = %v, want %v", ctx.InComment, tt.expectComm)
			}
		})
	}
}

func TestContextShouldFilter(t *testing.T) {
	tests := []struct {
		name     string
		ctx      Context
		expected bool
	}{
		{"test file", Context{InTest: true}, true},
		{"doc file", Context{InDocs: true}, true},
		{"example file", Context{InExample: true}, true},
		{"comment", Context{InComment: true}, true},
		{"regular source", Context{FileType: "source"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.ShouldFilter() != tt.expected {
				t.Errorf("ShouldFilter() = %v, want %v", tt.ctx.ShouldFilter(), tt.expected)
			}
		})
	}
}
