package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/NjariaOwen/sb-hub/pkg"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all sandboxes and archived data",
	Long:    `Displays a table of active Docker sandboxes and persistent data folders found on the host.`,
	Run: func(cmd *cobra.Command, args []string) {
		storageRoot := "/home/owen/prac-str"

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		defer cli.Close()

		ctx := context.Background()
		engine := &pkg.Dockerengine{Client: cli}

		activeMap, err := engine.GetActiveSandboxes(ctx)
		if err != nil {
			fmt.Printf("❌ Error fetching Docker status: %v\n", err)
			return
		}

		entries, err := os.ReadDir(storageRoot)
		if err != nil {
			fmt.Printf("❌ Error reading storage: %v\n", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
		fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tDOCKER ID\tSTORAGE PATH")

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			fullPath := filepath.Join(storageRoot, name)

			sandboxType := "Archived 💾"
			status := "Data Only"
			dockerID := "-"

			if c, exists := activeMap[name]; exists {
				sandboxType = "Active 🟢"
				status = c.State
				dockerID = c.ID[:12]
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, sandboxType, status, dockerID, fullPath)
		}
		w.Flush()
	},
}

func init() {

	rootCmd.AddCommand(listCmd)
}
