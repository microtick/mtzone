package microtick

import (
    "fmt"
    "os"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlocker(ctx sdk.Context, keeper Keeper) {
    // Monitor for end of chain
    params := keeper.GetParams(ctx)
    now := ctx.BlockHeader().Time
    
    if now.Unix() >= params.HaltTime {
   	    fmt.Printf("Reached prearranged chain end time: %s\n", time.Unix(params.HaltTime, 0).String())
   	    fmt.Println("Halting")
	    os.Exit(7)
    }
}
