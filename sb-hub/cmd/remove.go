package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove a sandbox",
	Long:    `Stops and removes a sandbox container. Can also wipe persistent data.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			name, _ = cmd.Flags().GetString("name")
		}

		if name == "" {
			fmt.Println("❌ Error: Name required. Use 'sb rm [name]'")
			return
		}

		forceAll, _ := cmd.Flags().GetBool("force-all")

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		defer cli.Close()

		engine := &pkg.Dockerengine{Client: cli}
		storagePath := filepath.Join("/home/owen/prac-str", name)

		shouldWipe := forceAll
		if !forceAll {
			fmt.Printf("❓ Wipe all data for '%s' or save it? [w]ipe / [s]ave: ", name)
			var choice string
			fmt.Scanln(&choice)
			if choice == "w" {
				shouldWipe = true
			}
		}

		ctx := context.Background()
		err = engine.RemoveSandbox(ctx, name, storagePath, shouldWipe)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Println("✅ Cleanup complete.")
		}
	},
}

func init() {
	removeCmd.Flags().StringP("name", "n", "", "Name of the sandbox to remove")
	removeCmd.Flags().Bool("force-all", false, "Wipe everything including persistent data")
	rootCmd.AddCommand(removeCmd)
}
