// Package main is the entry point for the zero CLI
package main

import (
	"fmt"
	"os"

	"github.com/crashappsec/zero/cmd/zero/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
