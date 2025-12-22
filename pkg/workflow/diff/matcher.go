// Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
// SPDX-License-Identifier: GPL-3.0

package diff

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Matcher handles finding comparison between scans
type Matcher struct {
	options DiffOptions
}

// NewMatcher creates a new finding matcher
func NewMatcher(options DiffOptions) *Matcher {
	return &Matcher{options: options}
}

// MatchFindings compares two sets of findings and returns match results
func (m *Matcher) MatchFindings(baseline, compare []FingerprintedFinding) []MatchResult {
	// Build maps for efficient lookup
	baselineByPrimary := make(map[string][]FingerprintedFinding)
	baselineByContent := make(map[string][]FingerprintedFinding)

	for _, f := range baseline {
		baselineByPrimary[f.Fingerprint.PrimaryKey] = append(baselineByPrimary[f.Fingerprint.PrimaryKey], f)
		if f.Fingerprint.ContentHash != "" {
			baselineByContent[f.Fingerprint.ContentHash] = append(baselineByContent[f.Fingerprint.ContentHash], f)
		}
	}

	var results []MatchResult
	matchedBaseline := make(map[string]bool) // Track matched baseline findings

	// Process each compare finding
	for _, cf := range compare {
		result := m.matchFinding(cf, baselineByPrimary, baselineByContent, matchedBaseline)
		results = append(results, result)
	}

	// Find unmatched baseline findings (fixed)
	for _, bf := range baseline {
		key := bf.Fingerprint.PrimaryKey + ":" + bf.Fingerprint.LocationKey
		if !matchedBaseline[key] {
			results = append(results, MatchResult{
				Status:      MatchFixed,
				OldFinding:  toPointer(toDeltaFinding(bf)),
				Confidence:  1.0,
				Explanation: "Finding no longer present in current scan",
			})
		}
	}

	return results
}

// matchFinding tries to find a matching baseline finding for a compare finding
func (m *Matcher) matchFinding(cf FingerprintedFinding, baselineByPrimary, baselineByContent map[string][]FingerprintedFinding, matched map[string]bool) MatchResult {
	// Try exact match first (same primary key and location)
	if baselines, ok := baselineByPrimary[cf.Fingerprint.PrimaryKey]; ok {
		for _, bf := range baselines {
			key := bf.Fingerprint.PrimaryKey + ":" + bf.Fingerprint.LocationKey
			if matched[key] {
				continue
			}

			if bf.Fingerprint.LocationKey == cf.Fingerprint.LocationKey {
				matched[key] = true
				return MatchResult{
					Status:      MatchExact,
					OldFinding:  toPointer(toDeltaFinding(bf)),
					NewFinding:  toPointer(toDeltaFinding(cf)),
					Confidence:  1.0,
					Explanation: "Exact match on primary key and location",
				}
			}
		}
	}

	// Try content hash match (code moved)
	if m.options.FuzzyMatch && cf.Fingerprint.ContentHash != "" {
		if baselines, ok := baselineByContent[cf.Fingerprint.ContentHash]; ok {
			for _, bf := range baselines {
				key := bf.Fingerprint.PrimaryKey + ":" + bf.Fingerprint.LocationKey
				if matched[key] {
					continue
				}

				// Same content but different location = moved
				if bf.Fingerprint.LocationKey != cf.Fingerprint.LocationKey {
					matched[key] = true
					return MatchResult{
						Status:      MatchMoved,
						OldFinding:  toPointer(toDeltaFinding(bf)),
						NewFinding:  toPointer(toDeltaFinding(cf)),
						Confidence:  0.95,
						Explanation: fmt.Sprintf("Content match, location changed from %s to %s", bf.Fingerprint.LocationKey, cf.Fingerprint.LocationKey),
					}
				}
			}
		}
	}

	// Try fuzzy location match (same primary key, different line within tolerance)
	if m.options.FuzzyMatch && m.options.LineTolerance > 0 {
		if baselines, ok := baselineByPrimary[cf.Fingerprint.PrimaryKey]; ok {
			for _, bf := range baselines {
				key := bf.Fingerprint.PrimaryKey + ":" + bf.Fingerprint.LocationKey
				if matched[key] {
					continue
				}

				if m.isLocationClose(bf.Fingerprint.LocationKey, cf.Fingerprint.LocationKey) {
					matched[key] = true
					return MatchResult{
						Status:      MatchSimilar,
						OldFinding:  toPointer(toDeltaFinding(bf)),
						NewFinding:  toPointer(toDeltaFinding(cf)),
						Confidence:  0.8,
						Explanation: fmt.Sprintf("Same primary key, location shifted from %s to %s", bf.Fingerprint.LocationKey, cf.Fingerprint.LocationKey),
					}
				}
			}
		}
	}

	// No match found - this is a new finding
	return MatchResult{
		Status:      MatchNew,
		NewFinding:  toPointer(toDeltaFinding(cf)),
		Confidence:  1.0,
		Explanation: "New finding not present in baseline",
	}
}

// isLocationClose checks if two location keys are within the line tolerance
func (m *Matcher) isLocationClose(loc1, loc2 string) bool {
	// Empty locations can't be compared
	if loc1 == "" || loc2 == "" {
		return false
	}

	// Parse locations (format: file:line or file:line:column)
	file1, line1 := parseLocation(loc1)
	file2, line2 := parseLocation(loc2)

	// Files must match
	if file1 != file2 {
		return false
	}

	// Check if lines are within tolerance
	diff := math.Abs(float64(line1 - line2))
	return diff <= float64(m.options.LineTolerance)
}

// parseLocation parses a location key into file and line
func parseLocation(loc string) (string, int) {
	parts := strings.Split(loc, ":")
	if len(parts) < 2 {
		return loc, 0
	}

	line, _ := strconv.Atoi(parts[1])
	return parts[0], line
}

// toDeltaFinding converts a FingerprintedFinding to DeltaFinding
func toDeltaFinding(f FingerprintedFinding) DeltaFinding {
	return DeltaFinding{
		Fingerprint: f.Fingerprint,
		Finding:     f.Finding,
		Severity:    f.Severity,
		Scanner:     f.Scanner,
		File:        f.File,
		Line:        f.Line,
		Message:     f.Message,
	}
}

// toPointer returns a pointer to a DeltaFinding
func toPointer(f DeltaFinding) *DeltaFinding {
	return &f
}

// FilterMatches filters match results based on options
func (m *Matcher) FilterMatches(results []MatchResult) []MatchResult {
	var filtered []MatchResult

	for _, r := range results {
		// Filter by show options
		if m.options.ShowNewOnly && r.Status != MatchNew {
			continue
		}
		if m.options.ShowFixedOnly && r.Status != MatchFixed {
			continue
		}

		// Filter by severity
		if len(m.options.Severities) > 0 {
			severity := getSeverity(r)
			if !containsIgnoreCase(m.options.Severities, severity) {
				continue
			}
		}

		// Filter out moved if not wanted
		if !m.options.IncludeMoved && r.Status == MatchMoved {
			continue
		}

		filtered = append(filtered, r)
	}

	return filtered
}

// getSeverity extracts severity from a match result
func getSeverity(r MatchResult) string {
	if r.NewFinding != nil {
		return r.NewFinding.Severity
	}
	if r.OldFinding != nil {
		return r.OldFinding.Severity
	}
	return ""
}

// containsIgnoreCase checks if a slice contains a string (case-insensitive)
func containsIgnoreCase(slice []string, s string) bool {
	s = strings.ToLower(s)
	for _, item := range slice {
		if strings.ToLower(item) == s {
			return true
		}
	}
	return false
}
