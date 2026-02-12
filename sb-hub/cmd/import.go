package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ComposeProject struct {
	Services map[string]struct {
		Image   string            `yaml:"image"`
		Volumes []string          `yaml:"volumes"`
		Ports   []string          `yaml:"ports"`
		Env     map[string]string `yaml:"environment"`
	} `yaml:"services"`
}

var importCmd = &cobra.Command{
	Use:   "import [path]",
	Short: "Import and sandbox a custom project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := filepath.Abs(args[0])
		projectName := filepath.Base(path)

		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()
		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		// 1. Dockerfile Logic
		dockerfilePath := filepath.Join(path, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); err == nil {
			tag := "sb-local-" + projectName
			if err := engine.BuildImage(ctx, path, tag); err != nil {
				fmt.Printf("❌ Build failed: %v\n", err)
				return
			}
			fmt.Printf("🚀 Launching custom sandbox: %s\n", projectName)
			createCmd.Flags().Set("image", tag)
			createCmd.Run(createCmd, []string{projectName})
			return
		}

		// 2. Docker Compose Logic
		composePath := filepath.Join(path, "docker-compose.yml")
		if _, err := os.Stat(composePath); err != nil {
			composePath = filepath.Join(path, "docker-compose.yaml")
		}

		if _, err := os.Stat(composePath); err == nil {
			fmt.Printf("🧩 Parsing Compose project: %s\n", projectName)
			data, err := ioutil.ReadFile(composePath)
			if err != nil {
				fmt.Printf("❌ Failed to read compose file: %v\n", err)
				return
			}

			var project ComposeProject
			if err := yaml.Unmarshal(data, &project); err != nil {
				fmt.Printf("❌ Failed to parse YAML: %v\n", err)
				return
			}

			engine.EnsureNetwork(ctx) // Ensure shared bridge

			for serviceName, config := range project.Services {
				uniqueName := fmt.Sprintf("%s-%s", projectName, serviceName)
				fmt.Printf("📦 Provisioning service: %s\n", uniqueName)

				createCmd.Flags().Set("image", config.Image)
				createCmd.Run(createCmd, []string{uniqueName})
			}
			return
		}

		fmt.Println("❌ No Dockerfile or docker-compose.yml found in path.")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
