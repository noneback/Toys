package main

import (
	"task/cmd"
	"task/db"
)

func main() {
	defer db.DB.Close()

	cmd.Execute()
}
