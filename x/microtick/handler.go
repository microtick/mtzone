package microtick

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	
	keeper "gitlab.com/microtick/mtzone/x/microtick/keeper"
	"gitlab.com/microtick/mtzone/x/microtick/msg"
)

func DenomChangeProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *msg.DenomChangeProposal:
			return handleDenomChangeProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized software upgrade proposal content type: %T", c)
		}
	}
}

func handleDenomChangeProposal(ctx sdk.Context, k keeper.Keeper, dcp *msg.DenomChangeProposal) error {
	logger := k.Logger(ctx)
	logger.Info("Denom Change Proposal:")
	logger.Info(fmt.Sprintf("  New External token type: %s", dcp.ExtDenom))
    logger.Info(fmt.Sprintf("  New Ratio: %d", dcp.ExtPerInt))
	
	// Clear markets
	k.ClearMarkets(ctx)
	
	k.SetExtTokenType(ctx, dcp.ExtDenom)
	k.SetExtPerInt(ctx, uint32(dcp.ExtPerInt))
	
    return nil
}
