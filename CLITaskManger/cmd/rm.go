package cmd

import (
	"fmt"
	"os"
	"strconv"
	"task/db"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rmCmd)
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "rm task",
	Long:  "romove task",
	Run: func(cmd *cobra.Command, args []string) {
		if db.DB == nil {
			fmt.Println("DB is nil")
			os.Exit(1)
		}
		for _, val := range args {
			id, err := strconv.Atoi(val)
			if err != nil {
				fmt.Println("no such task or type error")
				os.Exit(1)
			}
			db.DB.Update(db.RemoveTask(id))
			fmt.Printf("Task %v has been deleted\n", id)

		}
	},
}
