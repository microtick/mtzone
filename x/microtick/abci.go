package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "github.com/microtick/mtzone/x/microtick/keeper"
)

func EndBlocker(ctx sdk.Context, mtKeeper keeper.Keeper) {
    // Add commissions
    mtKeeper.Sweep(ctx)
}
