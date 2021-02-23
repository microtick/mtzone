package microtick

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	
	keeper "gitlab.com/microtick/mtzone/x/microtick/keeper"
	"gitlab.com/microtick/mtzone/x/microtick/msg"
)

func MicrotickProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *msg.DenomChangeProposal:
			return handleDenomChangeProposal(ctx, k, c)
		case *msg.AddMarketsProposal:
			return handleAddMarketsProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized software upgrade proposal content type: %T", c)
		}
	}
}

func handleDenomChangeProposal(ctx sdk.Context, k keeper.Keeper, dcp *msg.DenomChangeProposal) error {
	logger := k.Logger(ctx)
	logger.Info("Denom Change Proposal:")
	logger.Info(fmt.Sprintf("  Denom: %s", dcp.BackingDenom))
    logger.Info(fmt.Sprintf("  Ratio: %s", dcp.BackingRatio))
	
	// Clear markets
	k.ClearMarkets(ctx)
	
	k.SetBackingParams(ctx, dcp.BackingDenom, dcp.BackingRatio)
  return nil
}

func handleAddMarketsProposal(ctx sdk.Context, k keeper.Keeper, amp *msg.AddMarketsProposal) error {
	logger := k.Logger(ctx)
	logger.Info("Add Markets Proposal:")
	for i := range amp.Markets {
	  k.SetDataMarket(ctx, keeper.NewDataMarket(amp.Markets[i].Name, amp.Markets[i].Description))
      logger.Info(fmt.Sprintf("  %s: %s", amp.Markets[i].Name, amp.Markets[i].Description))
	}
	
  return nil
}
