package cmd

import (
	"fmt"
	"os"
	"strings"
	"task/db"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list tasks table",
	Long:  "list all the task haven't been done",
	Run: func(cmd *cobra.Command, args []string) {
		var flag string
		if len(args) == 0 || len(args) > 1 {
			fmt.Printf("NumOfParam is wrong\nUse default qurry flag: NotFinished\n\n")
			flag = "notfinished"
		} else {
			stat := strings.ToLower(strings.Trim(args[0], " "))
			if stat != "notfinished" && stat != "finished" && stat != "all" {
				fmt.Println("Flag error\nUse one of [NotFinished,Finished,All]")
				os.Exit(1)
			}

			flag = stat
		}

		db.DB.View(db.ShowList(flag))
	},
}
