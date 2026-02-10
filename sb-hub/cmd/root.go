package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sb",
	Short: "sb-hub is a CLI for managing development sandboxes",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
