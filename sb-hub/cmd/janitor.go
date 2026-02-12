package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var janitorCmd = &cobra.Command{
	Use:   "janitor",
	Short: "Start the background TTL enforcer",
	Run: func(cmd *cobra.Command, args []string) {
		once, _ := cmd.Flags().GetBool("once")
		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		engine := &pkg.Dockerengine{Client: cli}
		storageRoot := "/home/owen/prac-str"

		fmt.Println("🧹 Janitor service started. Monitoring TTLs...")

		for {
			ctx := context.Background()
			expired, err := engine.GetExpiredSandboxes(ctx)
			if err != nil {
				fmt.Printf("❌ Janitor Error: %v\n", err)
			}

			for _, c := range expired {
				name := filepath.Base(c.Names[0])
				fmt.Printf("⏰ TTL Expired for: %s. Archiving...\n", name)

				// 1. Stop and Remove Container
				engine.RemoveSandbox(ctx, name, "", false)

				// 2. Hybrid Move: Using sudo mv to handle root-owned container files
				oldPath := filepath.Join(storageRoot, name)
				newPath := filepath.Join(storageRoot, fmt.Sprintf("%s_janitor_%s", name, time.Now().Format("20060102150405")))

				fmt.Printf("📦 Archiving data to: %s\n", filepath.Base(newPath))
				err := exec.Command("sudo", "mv", oldPath, newPath).Run()
				if err != nil {
					fmt.Printf("❌ Failed to archive %s: %v\n", name, err)
				}
			}

			if once {
				break
			}
			time.Sleep(30 * time.Second)
		}
	},
}

func init() {
	janitorCmd.Flags().Bool("once", false, "Run one cleanup cycle and exit (useful for testing)")
	rootCmd.AddCommand(janitorCmd)
}
