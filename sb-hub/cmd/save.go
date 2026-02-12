package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save [name] [tag]",
	Args:  cobra.ExactArgs(2),
	Short: "Save a snapshot of a sandbox",
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		tag := args[1]
		storageRoot := "/home/owen/prac-str"
		src := filepath.Join(storageRoot, name)
		dst := filepath.Join(storageRoot, fmt.Sprintf("%s_snap_%s", name, tag))

		if _, err := os.Stat(src); os.IsNotExist(err) {
			fmt.Printf("❌ No active data found for '%s' to save.\n", name)
			return
		}

		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("🔄 Snapshot '%s' exists. Overwriting...\n", tag)
			exec.Command("sudo", "rm", "-rf", dst).Run()
		}

		fmt.Printf("📸 Saving snapshot of %s as %s...\n", name, tag)
		err := exec.Command("sudo", "cp", "-r", src, dst).Run()
		if err != nil {
			fmt.Printf("❌ Failed to save: %v\n", err)
			return
		}
		fmt.Println("✅ Saved successfully.")
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
