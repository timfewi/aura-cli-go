// Aura is an intelligent command-line interface assistant.
package main

import (
	"os"

	"github.com/timfewi/aura-cli-go/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
