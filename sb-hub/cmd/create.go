package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new sandbox",
	Long:  `Creates a new development sandbox with specified size and TTL.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			name, _ = cmd.Flags().GetString("name")
		}

		if name == "" {
			name = generateRandomName()
			fmt.Printf("🎲 No name provided. Generated: %s\n", name)
		}

		size, _ := cmd.Flags().GetString("size")
		ttlOverride, _ := cmd.Flags().GetDuration("ttl")

		spec, ok := pkg.SandboxSpecs[size]
		if !ok {
			fmt.Printf("❌ Error: '%s' is not a valid size. Use small, medium, large, or xlarge.\n", size)
			return
		}

		storageRoot := "/home/owen/prac-str/"
		sandboxPath := filepath.Join(storageRoot, name)

		if _, err := os.Stat(sandboxPath); err == nil {
			fmt.Printf("⚠️  Existing data found at %s\n", sandboxPath)
			fmt.Printf("Choose action: [a]ttach, [r]ename old, [c]ancel: ")

			var action string
			fmt.Scanln(&action)

			switch action {
			case "a":
				fmt.Println("🔗 Attaching existing data...")
			case "r":
				timestamp := time.Now().Format("20060102150405")
				oldPath := fmt.Sprintf("%s_old_%s", sandboxPath, timestamp)
				os.Rename(sandboxPath, oldPath)
				os.MkdirAll(sandboxPath, 0755)
				fmt.Printf("📦 Renamed old data to '%s'\n", filepath.Base(oldPath))
			default:
				fmt.Println("🛑 Operation cancelled")
				return
			}
		} else {
			os.MkdirAll(sandboxPath, 0755)
		}

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("❌ Error: Failed to initialize Docker client: %v\n", err)
			return
		}
		defer cli.Close()

		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		if err := engine.EnsureImage(ctx, spec.Image); err != nil {
			fmt.Printf("❌ Image Error: %v\n", err)
			return
		}

		config := &container.Config{
			Image: spec.Image,
		}

		hostConfig := &container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/data", sandboxPath),
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

		id, err := engine.CreateSandbox(ctx, name, finalTTL, config, hostConfig)
		if err != nil {
			fmt.Printf("❌ Create Error: %v\n", err)
			return
		}
		fmt.Printf("✅ Sandbox started! Name: %s ID: %s\n", name, id[:12])
	},
}

func generateRandomName() string {
	adjectives := []string{"swift", "brave", "cool", "mighty", "keen"}
	nouns := []string{"whale", "ship", "anchor", "pilot", "wave"}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s-%s-%d", adjectives[r.Intn(len(adjectives))], nouns[r.Intn(len(nouns))], r.Intn(1000))
}

func init() {
	createCmd.Flags().StringP("name", "n", "", "Name of the sandbox")
	createCmd.Flags().StringP("size", "s", "small", "Size (small, medium, large, xlarge)")
	createCmd.Flags().DurationP("ttl", "t", 0, "Time to live override (e.g., 1h, 30m)")
	rootCmd.AddCommand(createCmd)
}
