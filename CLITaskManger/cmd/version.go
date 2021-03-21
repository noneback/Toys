package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version of task manager",
	Long:  "Print the version of number of task manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Task manager , version of 1.0")
	},
}
