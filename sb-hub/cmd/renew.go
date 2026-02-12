package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var renewCmd = &cobra.Command{
	Use:   "renew [name] [duration]",
	Short: "Extend the TTL of an active sandbox",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		durationStr := args[1]

		extension, err := time.ParseDuration(durationStr)
		if err != nil {
			fmt.Printf("❌ Invalid duration: %v\n", err)
			return
		}

		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		inspect, err := engine.InspectSandbox(ctx, name)
		if err != nil {
			fmt.Printf("❌ Sandbox '%s' not found.\n", name)
			return
		}

		fmt.Printf("⏱️  Renewing %s for %s...\n", name, durationStr)

		engine.RemoveSandbox(ctx, name, "", false)

		id, err := engine.CreateSandbox(ctx, name, extension, inspect.Config.Labels["com.sbhub.size"], inspect.Config, inspect.HostConfig)
		if err != nil {
			fmt.Printf("❌ Renew failed: %v\n", err)
		} else {
			fmt.Printf("✅ Renewed. New ID: %s\n", id[:12])
		}
	},
}

func init() {
	rootCmd.AddCommand(renewCmd)
}
