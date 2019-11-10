package microtick

import (
    "fmt"
    "os"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

func EndBlocker(ctx sdk.Context, keeper keeper.MicrotickKeeper) {
    // Monitor for end of chain
    params := keeper.GetParams(ctx)
    now := ctx.BlockHeader().Time
    
    if now.Unix() >= params.HaltTime {
   	    fmt.Printf("Reached prearranged chain end time: %s\n", time.Unix(params.HaltTime, 0).String())
   	    fmt.Println("Halting")
	    os.Exit(7)
    }
}
