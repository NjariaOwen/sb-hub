package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Short: "View the output logs of a sandbox",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		follow, _ := cmd.Flags().GetBool("follow")

		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		_, err := engine.InspectSandbox(ctx, name)
		if err != nil {
			fmt.Printf("❌ Sandbox '%s' not found.\n", name)
			return
		}
		options := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     follow,
			Timestamps: true,
			Tail:       "50",
		}

		out, err := cli.ContainerLogs(ctx, name, options)
		if err != nil {
			fmt.Printf("❌ Failed to fetch logs: %v\n", err)
			return
		}
		defer out.Close()

		fmt.Printf("📋 Showing logs for %s...\n", name)
		io.Copy(os.Stdout, out)
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Stream live logs")
	rootCmd.AddCommand(logsCmd)
}
