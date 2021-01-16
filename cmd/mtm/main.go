package main

import (
	"os"
	
	"github.com/cosmos/cosmos-sdk/server"
  svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
  
	"gitlab.com/microtick/mtzone/app"
	"gitlab.com/microtick/mtzone/cmd/mtm/cmd"
)

// In main we set the custom version info and call the rootCmd
func main() {
	app.SetAppVersion()
	
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultHome); err != nil {
    switch e := err.(type) {
    case server.ErrorCode:
      os.Exit(e.Code)
    default:
      os.Exit(1)
    }
	}
}
