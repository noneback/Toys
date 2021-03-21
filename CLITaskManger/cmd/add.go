package cmd

import (
	"fmt"
	"os"
	"strings"
	"task/db"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add task",
	Long:  "add task and short in bolt",
	Run: func(cmd *cobra.Command, args []string) {
		if db.DB == nil {
			fmt.Println("DB is nil")
			os.Exit(1)
		}
		taskInfo := strings.Join(args, " ")
		db.DB.Update(db.AddTask(taskInfo))
		fmt.Println("add task:", taskInfo)
	},
}
