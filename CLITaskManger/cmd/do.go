package cmd

import (
	"fmt"
	"os"
	"strconv"
	"task/db"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doCmd)
}

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "finished task",
	Long:  "finished task and remove from bolt",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("missing param,needs an id num")
			os.Exit(1)
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("wrong params,needs an id num")
			os.Exit(1)
		}

		db.DB.Update(db.FinishTask(id))
		fmt.Printf("task %v finished\n", id)
	},
}
