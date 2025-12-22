// Package languages provides programming language detection utilities
package languages

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// DirectoryStats holds language statistics for a directory
type DirectoryStats struct {
	TotalFiles    int     `json:"total_files"`
	TotalLines    int     `json:"total_lines"`
	Languages     []Stats `json:"languages"`
	TopLanguage   string  `json:"top_language,omitempty"`
	LanguageCount int     `json:"language_count"`
}

// ScanOptions configures how the scanner works
type ScanOptions struct {
	IncludeVendored       bool     // Include vendored files
	IncludeGenerated      bool     // Include generated files
	IncludeDocumentation  bool     // Include documentation files
	OnlyProgramming       bool     // Only count programming languages
	ExcludeDirs           []string // Additional directories to skip
	MaxFiles              int      // Maximum files to scan (0 = unlimited)
}

// DefaultScanOptions returns sensible defaults for scanning
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		IncludeVendored:      false,
		IncludeGenerated:     false,
		IncludeDocumentation: false,
		OnlyProgramming:      true,
		ExcludeDirs: []string{
			".git", ".svn", ".hg",
		},
		MaxFiles: 0,
	}
}

// ScanDirectory scans a directory and returns language statistics
func ScanDirectory(root string, opts ScanOptions) (*DirectoryStats, error) {
	stats := &DirectoryStats{
		Languages: []Stats{},
	}

	langCounts := make(map[string]int)
	fileCount := 0

	excludeSet := make(map[string]bool)
	for _, d := range opts.ExcludeDirs {
		excludeSet[d] = true
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Get relative path for enry's vendor/generated detection
		relPath, _ := filepath.Rel(root, path)
		if relPath == "" {
			relPath = path
		}

		// Skip excluded directories
		if info.IsDir() {
			if excludeSet[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file limit
		if opts.MaxFiles > 0 && fileCount >= opts.MaxFiles {
			return filepath.SkipDir
		}

		// Skip vendored files (node_modules, vendor, etc.)
		if !opts.IncludeVendored && enry.IsVendor(relPath) {
			return nil
		}

		// Skip documentation files
		if !opts.IncludeDocumentation && enry.IsDocumentation(relPath) {
			return nil
		}

		// Detect language by filename first (fast path)
		lang, _ := enry.GetLanguageByFilename(info.Name())

		// If no match, try extension
		if lang == "" {
			lang, _ = enry.GetLanguageByExtension(info.Name())
		}

		if lang == "" {
			return nil
		}

		// Filter by language type if only programming languages
		if opts.OnlyProgramming {
			langType := enry.GetLanguageType(lang)
			if langType != enry.Programming {
				return nil
			}
		}

		langCounts[lang]++
		stats.TotalFiles++
		fileCount++

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Build language stats list
	for lang, count := range langCounts {
		percentage := 0.0
		if stats.TotalFiles > 0 {
			percentage = float64(count) / float64(stats.TotalFiles) * 100
		}
		langStats := Stats{
			Language:   lang,
			FileCount:  count,
			Percentage: percentage,
		}
		stats.Languages = append(stats.Languages, langStats)
	}

	// Sort by file count descending
	sort.Slice(stats.Languages, func(i, j int) bool {
		return stats.Languages[i].FileCount > stats.Languages[j].FileCount
	})

	stats.LanguageCount = len(stats.Languages)
	if len(stats.Languages) > 0 {
		stats.TopLanguage = stats.Languages[0].Language
	}

	return stats, nil
}

// TopLanguages returns the top N languages from stats
func TopLanguages(stats *DirectoryStats, n int) []Stats {
	if n <= 0 || n >= len(stats.Languages) {
		return stats.Languages
	}
	return stats.Languages[:n]
}

// FilterProgrammingLanguages returns only programming languages from stats
func FilterProgrammingLanguages(stats *DirectoryStats) []Stats {
	var result []Stats
	for _, s := range stats.Languages {
		if IsProgrammingLanguage(s.Language) {
			result = append(result, s)
		}
	}
	return result
}

// CountLines counts non-empty lines in a file
func CountLines(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	count := 0
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
