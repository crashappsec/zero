// Package languages provides programming language detection utilities
// using go-enry (a Go port of GitHub Linguist)
package languages

import (
	"os"
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
)

// Stats holds file counts and line counts for a language
type Stats struct {
	Language   string  `json:"language"`
	FileCount  int     `json:"file_count"`
	LineCount  int     `json:"line_count"`
	Percentage float64 `json:"percentage"`
}

// DetectFromPath returns the language for a file path based on extension or filename
func DetectFromPath(path string) string {
	filename := filepath.Base(path)
	lang, _ := enry.GetLanguageByFilename(filename)
	if lang != "" {
		return lang
	}
	lang, _ = enry.GetLanguageByExtension(filename)
	return lang
}

// DetectFromFile returns the language for a file, reading content if needed
func DetectFromFile(path string) string {
	filename := filepath.Base(path)

	// Try filename-based detection first (faster)
	if lang, _ := enry.GetLanguageByFilename(filename); lang != "" {
		return lang
	}

	// Read file content for more accurate detection
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// Use full detection with content
	return enry.GetLanguage(filename, content)
}

// DetectFromContent returns the language based on filename and content
func DetectFromContent(filename string, content []byte) string {
	return enry.GetLanguage(filename, content)
}

// IsProgrammingLanguage returns true if the language is a programming language
func IsProgrammingLanguage(lang string) bool {
	langType := enry.GetLanguageType(lang)
	return langType == enry.Programming
}

// IsMarkup returns true if the language is markup
func IsMarkup(lang string) bool {
	langType := enry.GetLanguageType(lang)
	return langType == enry.Markup
}

// IsData returns true if the language is a data format
func IsData(lang string) bool {
	langType := enry.GetLanguageType(lang)
	return langType == enry.Data
}

// IsProse returns true if the language is prose (documentation)
func IsProse(lang string) bool {
	langType := enry.GetLanguageType(lang)
	return langType == enry.Prose
}

// IsVendored returns true if the file path is a vendored file
func IsVendored(path string) bool {
	return enry.IsVendor(path)
}

// IsGenerated returns true if the file content appears to be generated
func IsGenerated(path string, content []byte) bool {
	return enry.IsGenerated(path, content)
}

// IsDocumentation returns true if the file path is documentation
func IsDocumentation(path string) bool {
	return enry.IsDocumentation(path)
}

// IsConfiguration returns true if the file path is a configuration file
func IsConfiguration(path string) bool {
	return enry.IsConfiguration(path)
}

// GetLanguageColor returns the color associated with a language (for UI)
func GetLanguageColor(lang string) string {
	return enry.GetColor(lang)
}

// GetLanguageExtensions returns the file extensions associated with a language
func GetLanguageExtensions(lang string) []string {
	return enry.GetLanguageExtensions(lang)
}
