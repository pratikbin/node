package application

import (
	"encoding/json"
	"log"

	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tendermintTypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *GaiaApp) ExportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs []string,
) (serverTypes.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, tendermintTypes.Header{Height: app.LastBlockHeight()})

	// We export at last height + 1, because that's the height at which
	// Tendermint will start InitChain.
	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		height = 0
		app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
	}

	genState := app.mm.ExportGenesis(ctx, app.appCodec)
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return serverTypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	return serverTypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, err
}

// prepare for fresh start at zero height
// NOTE zero height genesis is a temporary feature which will be deprecated
//      in favour of export at a block height
func (app *GaiaApp) prepForZeroHeightGenesis(ctx sdk.Context, jailAllowedAddrs []string) {
	applyAllowedAddrs := false

	// check if there is an allowed address list
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
	app.CrisisKeeper.AssertInvariants(ctx)

	/* Handle fee distribution state. */

	// withdraw all validator commission
	app.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingTypes.ValidatorI) (stop bool) {
		_, _ = app.DistributionKeeper.WithdrawValidatorCommission(ctx, val.GetOperator())
		return false
	})

	// withdraw all delegator rewards
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	for _, delegation := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		delAddr, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		_, _ = app.DistributionKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
	}

	// clear validator slash events
	app.DistributionKeeper.DeleteAllValidatorSlashEvents(ctx)

	// clear validator historical rewards
	app.DistributionKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	// set context height to zero
	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	// reinitialize all validators
	app.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingTypes.ValidatorI) (stop bool) {
		// donate any un-withdrawn outstanding reward fraction tokens to the community pool
		scraps := app.DistributionKeeper.GetValidatorOutstandingRewardsCoins(ctx, val.GetOperator())
		feePool := app.DistributionKeeper.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(scraps...)
		app.DistributionKeeper.SetFeePool(ctx, feePool)

		app.DistributionKeeper.Hooks().AfterValidatorCreated(ctx, val.GetOperator())
		return false
	})

	// reinitialize all delegations
	for _, del := range delegations {
		valAddr, err := sdk.ValAddressFromBech32(del.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delAddr, err := sdk.AccAddressFromBech32(del.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		app.DistributionKeeper.Hooks().BeforeDelegationCreated(ctx, delAddr, valAddr)
		app.DistributionKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr)
	}

	// reset context height
	ctx = ctx.WithBlockHeight(height)

	/* Handle staking state. */

	// iterate through redelegations, reset creation height
	app.StakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingTypes.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		app.StakingKeeper.SetRedelegation(ctx, red)
		return false
	})

	// iterate through unbonding delegations, reset creation height
	app.StakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingTypes.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
		return false
	})

	// Iterate through validators by power descending, reset bond heights, and
	// update bond intra-tx counters.
	store := ctx.KVStore(app.storeKeys[stakingTypes.StoreKey])
	iter := sdk.KVStoreReversePrefixIterator(store, stakingTypes.ValidatorsKey)
	counter := int16(0)

	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(stakingTypes.AddressFromValidatorsKey(iter.Key()))
		validator, found := app.StakingKeeper.GetValidator(ctx, addr)
		if !found {
			panic("expected validator, not found")
		}

		validator.UnbondingHeight = 0
		if applyAllowedAddrs && !allowedAddrsMap[addr.String()] {
			validator.Jailed = true
		}

		app.StakingKeeper.SetValidator(ctx, validator)
		counter++
	}

	err := iter.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		log.Fatal(err)
	}

	/* Handle slashing state. */

	// reset start height on signing infos
	app.SlashingKeeper.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info slashingTypes.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			app.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
			return false
		},
	)
}