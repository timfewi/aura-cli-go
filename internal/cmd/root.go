package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/timfewi/aura-cli-go/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "aura",
	Short: "Aura - Intelligent CLI Assistant",
	Long: `Aura is an intelligent command-line interface assistant designed to augment
your existing shell with context-aware suggestions, AI-powered assistance,
and intelligent navigation.`,
	Version: "1.0.0",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if err := config.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}
}
