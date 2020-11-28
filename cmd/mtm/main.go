package main

import (
	"os"

	"gitlab.com/microtick/mtzone/cmd/mtm/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		os.Exit(1)
	}
}
