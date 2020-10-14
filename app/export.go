package app

import (
	"encoding/json"
	"log"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	
  servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *MicrotickApp) ExportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs []string,
) (servertypes.ExportedApp, error) {

	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})

  height := app.LastBlockHeight() + 1
	if forZeroHeight {
		height = 0
		app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	}

	genState := app.mm.ExportGenesis(ctx, app.appCodec)
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators := staking.WriteValidators(ctx, app.keeper.staking)
	return servertypes.ExportedApp {
		AppState: appState,
		Validators: validators,
		Height: height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, nil
}

// prepare for fresh start at zero height
// NOTE zero height genesis is a temporary feature which will be deprecated
//      in favour of export at a block height
func (app *MicrotickApp) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) {
	applyAllowedAddrs := false

	//Check if there is a allowed address list
	if len(jailAllowedAddrs) > 0 {
		applyAllowedAddrs = true
	}

	allowedAddrsMap := make(map[string]bool)

	for _, addr := range jailAllowedAddrs {
		_, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			log.Fatal(err)
		}
		allowedAddrsMap[addr] = true
	}

	/* Just to be safe, assert the invariants on current state. */
	app.keeper.crisis.AssertInvariants(ctx)

	/* Handle fee distribution state. */

	// withdraw all validator commission
	app.keeper.staking.IterateValidators(ctx, func(_ int64, val exported.ValidatorI) (stop bool) {
		_, _ = app.keeper.distr.WithdrawValidatorCommission(ctx, val.GetOperator())
		return false
	})

	// withdraw all delegator rewards
	dels := app.keeper.staking.GetAllDelegations(ctx)
	for _, delegation := range dels {
		_, _ = app.keeper.distr.WithdrawDelegationRewards(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
	}

	// clear validator slash events
	app.keeper.distr.DeleteAllValidatorSlashEvents(ctx)

	// clear validator historical rewards
	app.keeper.distr.DeleteAllValidatorHistoricalRewards(ctx)

	// set context height to zero
	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	// reinitialize all validators
	app.keeper.staking.IterateValidators(ctx, func(_ int64, val exported.ValidatorI) (stop bool) {
		// donate any unwithdrawn outstanding reward fraction tokens to the community pool
		scraps := app.keeper.distr.GetValidatorOutstandingRewardsCoins(ctx, val.GetOperator())
		feePool := app.keeper.distr.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(scraps...)
		app.keeper.distr.SetFeePool(ctx, feePool)

		app.keeper.distr.Hooks().AfterValidatorCreated(ctx, val.GetOperator())
		return false
	})

	// reinitialize all delegations
	for _, del := range dels {
		app.keeper.distr.Hooks().BeforeDelegationCreated(ctx, del.DelegatorAddress, del.ValidatorAddress)
		app.keeper.distr.Hooks().AfterDelegationModified(ctx, del.DelegatorAddress, del.ValidatorAddress)
	}

	// reset context height
	ctx = ctx.WithBlockHeight(height)

	/* Handle staking state. */

	// iterate through redelegations, reset creation height
	app.keeper.staking.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		app.keeper.staking.SetRedelegation(ctx, red)
		return false
	})

	// iterate through unbonding delegations, reset creation height
	app.keeper.staking.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingtypes.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		app.keeper.staking.SetUnbondingDelegation(ctx, ubd)
		return false
	})

	// Iterate through validators by power descending, reset bond heights, and
	// update bond intra-tx counters.
	store := ctx.KVStore(app.keys[stakingtypes.StoreKey])
	iter := sdk.KVStoreReversePrefixIterator(store, stakingtypes.ValidatorsKey)
	counter := int16(0)

	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(iter.Key()[1:])
		validator, found := app.keeper.staking.GetValidator(ctx, addr)
		if !found {
			panic("expected validator, not found")
		}

		validator.UnbondingHeight = 0
		if applyAllowedAddrs && !allowedAddrsMap[addr.String()] {
			validator.Jailed = true
		}

		app.keeper.staking.SetValidator(ctx, validator)
		counter++
	}

	iter.Close()

	_ = app.keeper.staking.ApplyAndReturnValidatorSetUpdates(ctx)

	/* Handle slashing state. */

	// reset start height on signing infos
	app.keeper.slashing.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			app.keeper.slashing.SetValidatorSigningInfo(ctx, addr, info)
			return false
		},
	)
}
