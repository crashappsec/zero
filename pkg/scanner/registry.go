package scanner

import (
	"fmt"
	"sort"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Scanner)2
)

// Register adds a scanner to the registry
// This is typically called from scanner init() functions
func Register(s Scanner) { 
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[s.Name()] = s
}

// Get returns a scanner by name
func Get(name string) (Scanner, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	s, ok := registry[name]
	return s, ok
}

// List returns all registered scanner names
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetAll returns all registered scanners
func GetAll() []Scanner {
	registryMu.RLock()
	defer registryMu.RUnlock()

	scanners := make([]Scanner, 0, len(registry))
	for _, s := range registry {
		scanners = append(scanners, s)
	}
	return scanners
}

// GetByNames returns scanners for the given names
// Returns an error if any scanner is not found
func GetByNames(names []string) ([]Scanner, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	scanners := make([]Scanner, 0, len(names))
	for _, name := range names {
		s, ok := registry[name]
		if !ok {
			return nil, fmt.Errorf("scanner not found: %s", name)
		}
		scanners = append(scanners, s)
	}
	return scanners, nil
}

// TopologicalSort orders scanners by dependencies
// Scanners with no dependencies come first, then scanners that depend on them
func TopologicalSort(scanners []Scanner) ([]Scanner, error) {
	// Build dependency graph
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)
	scannerMap := make(map[string]Scanner)

	for _, s := range scanners {
		name := s.Name()
		scannerMap[name] = s
		inDegree[name] = 0
	}

	// Count incoming edges
	for _, s := range scanners {
		name := s.Name()
		for _, dep := range s.Dependencies() {
			// Only count dependencies that are in our scanner list
			if _, ok := scannerMap[dep]; ok {
				inDegree[name]++
				dependents[dep] = append(dependents[dep], name)
			}
		}
	}

	// Kahn's algorithm
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue) // Ensure deterministic ordering

	var result []Scanner
	for len(queue) > 0 {
		// Pop first element
		name := queue[0]
		queue = queue[1:]

		result = append(result, scannerMap[name])

		// Process dependents
		deps := dependents[name]
		sort.Strings(deps)
		for _, dep := range deps {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// Check for cycles
	if len(result) != len(scanners) {
		return nil, fmt.Errorf("dependency cycle detected in scanners")
	}

	return result, nil
}

// GroupByDependencies groups scanners into levels based on dependencies
// Scanners in the same level can run in parallel
func GroupByDependencies(scanners []Scanner) ([][]Scanner, error) {
	sorted, err := TopologicalSort(scanners)
	if err != nil {
		return nil, err
	}

	// Build completed set as we process
	completed := make(map[string]bool)
	var levels [][]Scanner
	var currentLevel []Scanner

	for _, s := range sorted {
		// Check if all dependencies are in previous levels
		allDepsSatisfied := true
		for _, dep := range s.Dependencies() {
			if !completed[dep] {
				// Dependency not yet completed, check if it's even in our list
				found := false
				for _, other := range scanners {
					if other.Name() == dep {
						found = true
						break
					}
				}
				if found {
					allDepsSatisfied = false
					break
				}
			}
		}

		if allDepsSatisfied {
			currentLevel = append(currentLevel, s)
		} else {
			// Start new level
			if len(currentLevel) > 0 {
				levels = append(levels, currentLevel)
				// Mark all in current level as completed
				for _, cs := range currentLevel {
					completed[cs.Name()] = true
				}
			}
			currentLevel = []Scanner{s}
		}
	}

	// Don't forget the last level
	if len(currentLevel) > 0 {
		levels = append(levels, currentLevel)
	}

	return levels, nil
}

// Clear removes all registered scanners (useful for testing)
func Clear() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]Scanner)
}
