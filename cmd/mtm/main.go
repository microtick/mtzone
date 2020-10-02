package main

import (
	"os"

	"gitlab.com/microtick/mtzone/app"
	"gitlab.com/microtick/mtzone/cmd/mtm/cmd"
)

// In main we set the custom version info and call the rootCmd
func main() {
	app.SetAppVersion()
	
	rootCmd, _ := cmd.NewRootCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		os.Exit(1)
	}
}
