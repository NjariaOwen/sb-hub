package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var consoleCmd = &cobra.Command{
	Use:     "console [name]",
	Aliases: []string{"enter", "shell", "exec"},
	Short:   "Open an interactive terminal inside a sandbox",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		inspect, err := engine.InspectSandbox(ctx, name)
		if err != nil {
			fmt.Printf("❌ Sandbox '%s' not found.\n", name)
			return
		}

		if !inspect.State.Running {
			fmt.Printf("❌ Sandbox '%s' is %s. Please start it first.\n", name, inspect.State.Status)
			return
		}

		fmt.Printf("🔌 Connecting to %s console... (type 'exit' to disconnect)\n", name)

		shellCmd := exec.Command("docker", "exec", "-it", name, "/bin/sh")

		shellCmd.Stdin = os.Stdin
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr

		err = shellCmd.Run()
		if err != nil {
			fmt.Printf("❌ Console session ended with error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(consoleCmd)
}
