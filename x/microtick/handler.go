package microtick

import (
    "fmt"
    "os"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

func EndBlocker(ctx sdk.Context, mtKeeper keeper.Keeper) {
    // Reward dust
    dust := mtKeeper.Sweep(ctx)
    fmt.Printf("Pool: %s\n", dust.String())
    
    // Monitor for end of chain
    haltTime := mtKeeper.GetHaltTime(ctx)
    now := ctx.BlockHeader().Time.UTC().Unix()
    
    //fmt.Printf("Halt Time: %d\n", haltTime)
    //fmt.Printf("Now:       %d\n", now)
    
    if now >= haltTime {
   	    fmt.Printf("Reached prearranged chain end time: %s\n", time.Unix(haltTime, 0).UTC().Format(mt.TimeFormat))
   	    fmt.Println("Halting")
	    os.Exit(7)
    }
}
