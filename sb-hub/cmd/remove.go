package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove sandboxes and their volumes",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			name, _ = cmd.Flags().GetString("name")
		}
		if name == "" {
			fmt.Println("❌ Name required")
			return
		}

		volOnly, _ := cmd.Flags().GetBool("vol-only")
		storagePath := filepath.Join("/home/owen/prac-str", name)

		if strings.Contains(name, "_snap_") || volOnly {
			fmt.Printf("🧹 Wiping volume data at %s...\n", storagePath)
			if err := exec.Command("sudo", "rm", "-rf", storagePath).Run(); err != nil {
				fmt.Printf("❌ Failed to wipe folder: %v\n", err)
				return
			}
			fmt.Println("✅ Volume data removed.")
			return
		}

		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		fmt.Printf("🗑️  Removing sandbox %s and all associated data...\n", name)

		engine.RemoveSandbox(ctx, name, "", false)

		if err := exec.Command("sudo", "rm", "-rf", storagePath).Run(); err != nil {
			fmt.Printf("❌ Failed to wipe storage path: %v\n", err)
		} else {
			fmt.Println("✅ Done.")
		}
	},
}

func init() {
	removeCmd.Flags().StringP("name", "n", "", "Name of the sandbox")
	removeCmd.Flags().Bool("vol-only", false, "Wipe only the data folder")
	rootCmd.AddCommand(removeCmd)
}
