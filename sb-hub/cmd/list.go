package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all sandboxes and archived data",
	Run: func(cmd *cobra.Command, args []string) {
		storageRoot := "/home/owen/prac-str"
		cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer cli.Close()

		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}
		activeMap, _ := engine.GetActiveSandboxes(ctx)

		entries, _ := os.ReadDir(storageRoot)
		// We add PORT to the header
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "NAME\tTYPE\tSIZE\tSTATUS\tIMAGE\tPORT\tTTL REMAINING\tSTORAGE PATH")

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			fullPath := filepath.Join(storageRoot, name)

			sandboxType := "Archived 💾"
			size := "-"
			status := "Data Only"
			imageName := "-"
			port := "-" // Default for archived data
			ttlRemaining := "-"

			if c, exists := activeMap[name]; exists {
				sandboxType = "Active 🟢"
				status = c.State
				imageName = c.Image
				size = c.Labels["com.sbhub.size"]

				// Pull the host port from labels
				if p, ok := c.Labels["com.sbhub.hostport"]; ok {
					port = p
				}

				if exp, ok := c.Labels["com.sbhub.expires"]; ok {
					t, err := time.Parse(time.RFC3339, exp)
					if err == nil {
						rem := time.Until(t).Round(time.Second)
						if rem > 0 {
							ttlRemaining = rem.String()
						} else {
							ttlRemaining = "EXPIRED"
						}
					}
				}
			}
			// Added port variable to the output string
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", name, sandboxType, size, status, imageName, port, ttlRemaining, fullPath)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
