package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
)

func FindFreePort(start, end int, used map[string]bool) int {
	for port := start; port <= end; port++ {
		pStr := fmt.Sprintf("%d", port)
		if used[pStr] {
			continue
		}
		ln, err := net.Listen("tcp", ":"+pStr)
		if err == nil {
			ln.Close()
			return port
		}
	}
	return 0
}

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a networked sandbox with auto-port mapping",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			name, _ = cmd.Flags().GetString("name")
		}
		if name == "" {
			name = pkg.GenerateRandomName()
		}

		size, _ := cmd.Flags().GetString("size")
		customImg, _ := cmd.Flags().GetString("image")
		restoreTag, _ := cmd.Flags().GetString("restore")
		ttlOverride, _ := cmd.Flags().GetDuration("ttl")

		spec, ok := pkg.SandboxSpecs[size]
		if !ok {
			fmt.Printf("❌ Invalid size: %s\n", size)
			return
		}

		imageToUse := spec.Image
		if customImg != "" {
			imageToUse = customImg
		}

		storageRoot := "/home/owen/prac-str/"
		sandboxPath := filepath.Join(storageRoot, name)

		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		// 1. Networking and Port Logic
		engine.EnsureNetwork(ctx)
		usedPorts, _ := engine.GetUsedPorts(ctx)
		hostPort := FindFreePort(8000, 9000, usedPorts)

		if restoreTag != "" {
			snapPath := filepath.Join(storageRoot, restoreTag)
			if _, err := os.Stat(snapPath); os.IsNotExist(err) {
				snapPath = filepath.Join(storageRoot, fmt.Sprintf("%s_snap_%s", name, restoreTag))
			}
			if _, err := os.Stat(snapPath); err == nil {
				fmt.Printf("🔄 Restoring data from: %s\n", snapPath)
				exec.Command("sudo", "rm", "-rf", sandboxPath).Run()
				exec.Command("sudo", "cp", "-r", snapPath, sandboxPath).Run()
			}
		}

		if _, err := os.Stat(sandboxPath); err == nil && restoreTag == "" {
			fmt.Printf("⚠️  Existing data found. [a]ttach, [r]ename, [c]ancel: ")
			var action string
			fmt.Scanln(&action)
			if action == "r" {
				oldPath := fmt.Sprintf("%s_old_%s", sandboxPath, time.Now().Format("20060102150405"))
				os.Rename(sandboxPath, oldPath)
				engine.RemoveSandbox(ctx, name, "", false)
			} else if action == "a" {
				engine.RemoveSandbox(ctx, name, "", false)
			} else {
				return
			}
		}
		os.MkdirAll(sandboxPath, 0755)

		engine.EnsureImage(ctx, imageToUse)

		hostConfig := &container.HostConfig{
			Binds:       []string{fmt.Sprintf("%s:/data", sandboxPath)},
			NetworkMode: "sb-hub-net",
			PortBindings: nat.PortMap{
				"80/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", hostPort)}},
			},
			Resources: container.Resources{
				NanoCPUs: int64(spec.CPUCores * 1e9),
				Memory:   int64(spec.MemoryMB * 1024 * 1024),
			},
		}

		finalTTL := ttlOverride
		if finalTTL == 0 {
			finalTTL = spec.DefaultTTL
		}

		config := &container.Config{
			Image: imageToUse,
			Labels: map[string]string{
				"com.sbhub.hostport": fmt.Sprintf("%d", hostPort),
			},
		}

		id, err := engine.CreateSandbox(ctx, name, finalTTL, size, config, hostConfig)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			// FIXED: Now using 'id' to satisfy the Go compiler
			fmt.Printf("✅ Started %s (ID: %s) at http://localhost:%d\n", name, id[:12], hostPort)
		}
	},
}

func init() {
	createCmd.Flags().StringP("name", "n", "", "Sandbox name")
	createCmd.Flags().StringP("size", "s", "small", "Size preset")
	createCmd.Flags().StringP("image", "i", "", "Custom Docker image")
	createCmd.Flags().StringP("restore", "r", "", "Snapshot folder or tag to restore")
	createCmd.Flags().DurationP("ttl", "t", 0, "TTL override")
	rootCmd.AddCommand(createCmd)
}
