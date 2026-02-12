package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var detachCmd = &cobra.Command{
	Use:   "detach [name]",
	Short: "Remove storage mounts from a sandbox",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		inspect, _ := engine.InspectSandbox(ctx, name)
		fmt.Printf("🔌 Making %s stateless...\n", name)

		engine.RemoveSandbox(ctx, name, "", false)
		inspect.HostConfig.Binds = nil

		engine.CreateSandbox(ctx, name, 1*time.Hour, inspect.Config.Labels["com.sbhub.size"], inspect.Config, inspect.HostConfig)
		fmt.Println("✅ Detached.")
	},
}

func init() { rootCmd.AddCommand(detachCmd) }
