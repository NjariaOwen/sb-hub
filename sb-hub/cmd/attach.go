package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach [sandbox] [name]",
	Short: "Switch a sandbox to a different data folder",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		name, folder := args[0], args[1]
		newPath := filepath.Join("/home/owen/prac-str", folder)

		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		inspect, err := engine.InspectSandbox(ctx, name)
		if err != nil {
			fmt.Printf("❌ Sandbox '%s' not found\n", name)
			return
		}

		fmt.Printf("🔄 Attaching sandbox '%s' to folder '%s'\n", name, folder)
		engine.RemoveSandbox(ctx, name, "", false)

		inspect.HostConfig.Binds = []string{fmt.Sprintf("%s:/data", newPath)}

		id, err := engine.CreateSandbox(ctx, name, 1*time.Hour, inspect.Config.Labels["com.sbhub.size"], inspect.Config, inspect.HostConfig)
		if err == nil {
			fmt.Printf("✅ Attached. New ID: %s\n", id[:12])
		}

	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
