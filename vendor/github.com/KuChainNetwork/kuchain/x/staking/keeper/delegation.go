package keeper

import (
	"bytes"
	"fmt"
	"time"

	"github.com/KuChainNetwork/kuchain/chain/constants"
	"github.com/KuChainNetwork/kuchain/chain/store"
	stakingexport "github.com/KuChainNetwork/kuchain/x/staking/exported"
	"github.com/KuChainNetwork/kuchain/x/staking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// return a specific delegation
func (k Keeper) GetDelegation(ctx sdk.Context,
	delAddr AccountID, valAddr AccountID) (
	delegation types.Delegation, found bool) {

	store := store.NewStore(ctx, k.storeKey)
	key := types.GetDelegationKey(delAddr, valAddr)
	value := store.Get(key)
	if value == nil {
		return delegation, false
	}

	delegation = types.MustUnmarshalDelegation(k.cdc, value)
	return delegation, true
}

// IterateAllDelegations iterate through all of the delegations
func (k Keeper) IterateAllDelegations(ctx sdk.Context, cb func(delegation types.Delegation) (stop bool)) {
	store := store.NewStore(ctx, k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if cb(delegation) {
			break
		}
	}
}

// GetAllDelegations returns all delegations used during genesis dump
func (k Keeper) GetAllDelegations(ctx sdk.Context) (delegations []types.Delegation) {
	k.IterateAllDelegations(ctx, func(delegation types.Delegation) bool {
		delegations = append(delegations, delegation)
		return false
	})
	return delegations
}

// return all delegations to a specific validator. Useful for querier.
func (k Keeper) GetValidatorDelegations(ctx sdk.Context, valAddr AccountID) (delegations []types.Delegation) { //nolint:interfacer
	store := store.NewStore(ctx, k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		if delegation.GetValidatorAccountID().Eq(valAddr) {
			delegations = append(delegations, delegation)
		}
	}
	return delegations
}

// return a given amount of all the delegations from a delegator
func (k Keeper) GetDelegatorDelegations(ctx sdk.Context, delegator AccountID,
	maxRetrieve uint16) (delegations []types.Delegation) {

	delegations = make([]types.Delegation, maxRetrieve)

	store := store.NewStore(ctx, k.storeKey)
	delegatorPrefixKey := types.GetDelegationsKey(delegator)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		delegation := types.MustUnmarshalDelegation(k.cdc, iterator.Value())
		delegations[i] = delegation
		i++
	}
	return delegations[:i] // trim if the array length < maxRetrieve
}

// set a delegation
func (k Keeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := store.NewStore(ctx, k.storeKey)
	b := types.MustMarshalDelegation(k.cdc, delegation)
	store.Set(types.GetDelegationKey(delegation.DelegatorAccount, delegation.ValidatorAccount), b)
}

// remove a delegation
func (k Keeper) RemoveDelegation(ctx sdk.Context, delegation types.Delegation) {
	// TODO: Consider calling hooks outside of the store wrapper functions, it's unobvious.
	k.BeforeDelegationRemoved(ctx, delegation.ValidatorAccount, delegation.DelegatorAccount)
	store := store.NewStore(ctx, k.storeKey)
	store.Delete(types.GetDelegationKey(delegation.DelegatorAccount, delegation.ValidatorAccount))
}

// return a given amount of all the delegator unbonding-delegations
func (k Keeper) GetUnbondingDelegations(ctx sdk.Context, delegator AccountID,
	maxRetrieve uint16) (unbondingDelegations []types.UnbondingDelegation) {

	unbondingDelegations = make([]types.UnbondingDelegation, maxRetrieve)

	store := store.NewStore(ctx, k.storeKey)
	delegatorPrefixKey := types.GetUBDsKey(delegator.StoreKey())
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		unbondingDelegation := types.MustUnmarshalUBD(k.cdc, iterator.Value())
		unbondingDelegations[i] = unbondingDelegation
		i++
	}
	return unbondingDelegations[:i] // trim if the array length < maxRetrieve
}

// return a unbonding delegation
func (k Keeper) GetUnbondingDelegation(
	ctx sdk.Context, delAddr AccountID, valAddr AccountID,
) (ubd types.UnbondingDelegation, found bool) {

	store := store.NewStore(ctx, k.storeKey)
	key := types.GetUBDKey(delAddr.StoreKey(), valAddr.StoreKey())
	value := store.Get(key)
	if value == nil {
		return ubd, false
	}

	ubd = types.MustUnmarshalUBD(k.cdc, value)
	return ubd, true
}

// return all unbonding delegations from a particular validator
func (k Keeper) GetUnbondingDelegationsFromValidator(ctx sdk.Context, valAddr AccountID) (ubds []types.UnbondingDelegation) {
	store := store.NewStore(ctx, k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetUBDsByValIndexKey(valAddr))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := types.GetUBDKeyFromValIndexKey(iterator.Key())
		value := store.Get(key)
		ubd := types.MustUnmarshalUBD(k.cdc, value)
		ubds = append(ubds, ubd)
	}
	return ubds
}

// iterate through all of the unbonding delegations
func (k Keeper) IterateUnbondingDelegations(ctx sdk.Context, fn func(index int64, ubd types.UnbondingDelegation) (stop bool)) {
	store := store.NewStore(ctx, k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.UnbondingDelegationKey)
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(k.cdc, iterator.Value())
		if stop := fn(i, ubd); stop {
			break
		}
		i++
	}
}

// HasMaxUnbondingDelegationEntries - check if unbonding delegation has maximum number of entries
func (k Keeper) HasMaxUnbondingDelegationEntries(ctx sdk.Context,
	delegatorAddr AccountID, validatorAddr AccountID) bool {

	name, ok := delegatorAddr.ToName()
	if ok && constants.IsSystemAccount(name) {
		return false
	}

	ubd, found := k.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		return false
	}
	return len(ubd.Entries) >= int(k.MaxEntries(ctx))
}

// set the unbonding delegation and associated index
func (k Keeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := store.NewStore(ctx, k.storeKey)
	bz := types.MustMarshalUBD(k.cdc, ubd)
	key := types.GetUBDKey(ubd.DelegatorAccount.StoreKey(), ubd.ValidatorAccount.StoreKey())
	store.Set(key, bz)
	store.Set(types.GetUBDByValIndexKey(ubd.DelegatorAccount, ubd.ValidatorAccount), bz) // index, store empty bytes
}

// remove the unbonding delegation object and associated index
func (k Keeper) RemoveUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) {
	store := store.NewStore(ctx, k.storeKey)
	key := types.GetUBDKey(ubd.DelegatorAccount.StoreKey(), ubd.ValidatorAccount.StoreKey())
	store.Delete(key)
	store.Delete(types.GetUBDByValIndexKey(ubd.DelegatorAccount, ubd.ValidatorAccount))
}

// SetUnbondingDelegationEntry adds an entry to the unbonding delegation at
// the given addresses. It creates the unbonding delegation if it does not exist
func (k Keeper) SetUnbondingDelegationEntry(
	ctx sdk.Context, delegatorAddr AccountID, validatorAddr AccountID,
	creationHeight int64, minTime time.Time, balance sdk.Int,
) types.UnbondingDelegation {

	ubd, found := k.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	if found {
		ubd.AddEntry(creationHeight, minTime, balance)
	} else {
		ubd = types.NewUnbondingDelegation(delegatorAddr, validatorAddr, creationHeight, minTime, balance)
	}

	k.SetUnbondingDelegation(ctx, ubd)
	return ubd
}

// unbonding delegation queue timeslice operations

// gets a specific unbonding queue timeslice. A timeslice is a slice of DVPairs
// corresponding to unbonding delegations that expire at a certain time.
func (k Keeper) GetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (dvPairs []types.DVPair) {
	store := store.NewStore(ctx, k.storeKey)
	bz := store.Get(types.GetUnbondingDelegationTimeKey(timestamp))
	if bz == nil {
		return []types.DVPair{}
	}

	pairs := types.DVPairs{}
	k.cdc.MustUnmarshalBinaryBare(bz, &pairs)
	return pairs.Pairs
}

// Sets a specific unbonding queue timeslice.
func (k Keeper) SetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []types.DVPair) {
	store := store.NewStore(ctx, k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&types.DVPairs{Pairs: keys})
	store.Set(types.GetUnbondingDelegationTimeKey(timestamp), bz)
}

// Insert an unbonding delegation to the appropriate timeslice in the unbonding queue
func (k Keeper) InsertUBDQueue(ctx sdk.Context, ubd types.UnbondingDelegation,
	completionTime time.Time) {

	timeSlice := k.GetUBDQueueTimeSlice(ctx, completionTime)
	dvPair := types.DVPair{DelegatorAccount: ubd.DelegatorAccount, ValidatorAccount: ubd.ValidatorAccount}
	if len(timeSlice) == 0 {
		k.SetUBDQueueTimeSlice(ctx, completionTime, []types.DVPair{dvPair})
	} else {
		timeSlice = append(timeSlice, dvPair)
		k.SetUBDQueueTimeSlice(ctx, completionTime, timeSlice)
	}
}

// Returns all the unbonding queue timeslices from time 0 until endTime
func (k Keeper) UBDQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := store.NewStore(ctx, k.storeKey)
	return store.Iterator(types.UnbondingQueueKey,
		sdk.InclusiveEndBytes(types.GetUnbondingDelegationTimeKey(endTime)))
}

// Returns a concatenated list of all the timeslices inclusively previous to
// currTime, and deletes the timeslices from the queue
func (k Keeper) DequeueAllMatureUBDQueue(ctx sdk.Context, currTime time.Time) (matureUnbonds []types.DVPair) {
	store := store.NewStore(ctx, k.storeKey)

	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	unbondingTimesliceIterator := k.UBDQueueIterator(ctx, ctx.BlockHeader().Time)
	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := types.DVPairs{}
		value := unbondingTimesliceIterator.Value()
		k.cdc.MustUnmarshalBinaryBare(value, &timeslice)

		matureUnbonds = append(matureUnbonds, timeslice.Pairs...)
		store.Delete(unbondingTimesliceIterator.Key())
	}

	return matureUnbonds
}

// return a given amount of all the delegator redelegations
func (k Keeper) GetRedelegations(ctx sdk.Context, delegator AccountID,
	maxRetrieve uint16) (redelegations []types.Redelegation) {
	redelegations = make([]types.Redelegation, maxRetrieve)

	store := store.NewStore(ctx, k.storeKey)
	delegatorPrefixKey := types.GetREDsKey(delegator.StoreKey())
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		redelegation := types.MustUnmarshalRED(k.cdc, iterator.Value())
		redelegations[i] = redelegation
		i++
	}
	return redelegations[:i] // trim if the array length < maxRetrieve
}

// return a redelegation
func (k Keeper) GetRedelegation(ctx sdk.Context,
	delAddr AccountID, valSrcAddr, valDstAddr AccountID) (red types.Redelegation, found bool) {

	store := store.NewStore(ctx, k.storeKey)
	key := types.GetREDKey(delAddr.StoreKey(), valSrcAddr.StoreKey(), valDstAddr.StoreKey())
	value := store.Get(key)
	if value == nil {
		return red, false
	}

	red = types.MustUnmarshalRED(k.cdc, value)
	return red, true
}

// return all redelegations from a particular validator
func (k Keeper) GetRedelegationsFromSrcValidator(ctx sdk.Context, valAddr AccountID) (reds []types.Redelegation) {
	store := store.NewStore(ctx, k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetREDsFromValSrcIndexKey(valAddr.StoreKey()))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := types.GetREDKeyFromValSrcIndexKey(iterator.Key())
		value := store.Get(key)
		red := types.MustUnmarshalRED(k.cdc, value)
		reds = append(reds, red)
	}
	return reds
}

// check if validator is receiving a redelegation
func (k Keeper) HasReceivingRedelegation(ctx sdk.Context,
	delAddr AccountID, valDstAddr AccountID) bool {

	name, ok := delAddr.ToName()
	if ok && constants.IsSystemAccount(name) {
		return false
	}
	store := store.NewStore(ctx, k.storeKey)
	prefix := types.GetREDsByDelToValDstIndexKey(delAddr, valDstAddr)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	return iterator.Valid()
}

// HasMaxRedelegationEntries - redelegation has maximum number of entries
func (k Keeper) HasMaxRedelegationEntries(ctx sdk.Context,
	delegatorAddr AccountID, validatorSrcAddr,
	validatorDstAddr AccountID) bool {

	name, ok := delegatorAddr.ToName()
	if ok && constants.IsSystemAccount(name) {
		return false
	}

	red, found := k.GetRedelegation(ctx, delegatorAddr, validatorSrcAddr, validatorDstAddr)
	if !found {
		return false
	}
	return len(red.Entries) >= int(k.MaxEntries(ctx))
}

// set a redelegation and associated index
func (k Keeper) SetRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := store.NewStore(ctx, k.storeKey)
	bz := types.MustMarshalRED(k.cdc, red)
	key := types.GetREDKey(red.DelegatorAccount.StoreKey(), red.ValidatorSrcAccount.StoreKey(), red.ValidatorDstAccount.StoreKey())
	store.Set(key, bz)
	store.Set(types.GetREDByValSrcIndexKey(red.DelegatorAccount.StoreKey(), red.ValidatorSrcAccount.StoreKey(), red.ValidatorDstAccount.StoreKey()), bz)
	store.Set(types.GetREDByValDstIndexKey(red.DelegatorAccount, red.ValidatorSrcAccount, red.ValidatorDstAccount), bz)
}

// SetUnbondingDelegationEntry adds an entry to the unbonding delegation at
// the given addresses. It creates the unbonding delegation if it does not exist
func (k Keeper) SetRedelegationEntry(ctx sdk.Context,
	delegatorAddr AccountID, validatorSrcAddr,
	validatorDstAddr AccountID, creationHeight int64,
	minTime time.Time, balance sdk.Int,
	sharesSrc, sharesDst sdk.Dec) types.Redelegation {

	red, found := k.GetRedelegation(ctx, delegatorAddr, validatorSrcAddr, validatorDstAddr)
	if found {
		red.AddEntry(creationHeight, minTime, balance, sharesDst)
	} else {
		red = types.NewRedelegation(delegatorAddr, validatorSrcAddr,
			validatorDstAddr, creationHeight, minTime, balance, sharesDst)
	}
	k.SetRedelegation(ctx, red)
	return red
}

// iterate through all redelegations
func (k Keeper) IterateRedelegations(ctx sdk.Context, fn func(index int64, red types.Redelegation) (stop bool)) {
	store := store.NewStore(ctx, k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RedelegationKey)
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		red := types.MustUnmarshalRED(k.cdc, iterator.Value())
		if stop := fn(i, red); stop {
			break
		}
		i++
	}
}

// remove a redelegation object and associated index
func (k Keeper) RemoveRedelegation(ctx sdk.Context, red types.Redelegation) {
	store := store.NewStore(ctx, k.storeKey)
	redKey := types.GetREDKey(red.DelegatorAccount.StoreKey(), red.ValidatorSrcAccount.StoreKey(), red.ValidatorDstAccount.StoreKey())
	store.Delete(redKey)
	store.Delete(types.GetREDByValSrcIndexKey(red.DelegatorAccount.StoreKey(), red.ValidatorSrcAccount.StoreKey(), red.ValidatorDstAccount.StoreKey()))
	store.Delete(types.GetREDByValDstIndexKey(red.DelegatorAccount, red.ValidatorSrcAccount, red.ValidatorDstAccount))
}

// redelegation queue timeslice operations

// Gets a specific redelegation queue timeslice. A timeslice is a slice of DVVTriplets corresponding to redelegations
// that expire at a certain time.
func (k Keeper) GetRedelegationQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (dvvTriplets []types.DVVTriplet) {
	store := store.NewStore(ctx, k.storeKey)
	bz := store.Get(types.GetRedelegationTimeKey(timestamp))
	if bz == nil {
		return []types.DVVTriplet{}
	}

	triplets := types.DVVTriplets{}
	k.cdc.MustUnmarshalBinaryBare(bz, &triplets)
	return triplets.Triplets
}

// Sets a specific redelegation queue timeslice.
func (k Keeper) SetRedelegationQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []types.DVVTriplet) {
	store := store.NewStore(ctx, k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&types.DVVTriplets{Triplets: keys})
	store.Set(types.GetRedelegationTimeKey(timestamp), bz)
}

// Insert an redelegation delegation to the appropriate timeslice in the redelegation queue
func (k Keeper) InsertRedelegationQueue(ctx sdk.Context, red types.Redelegation,
	completionTime time.Time) {

	timeSlice := k.GetRedelegationQueueTimeSlice(ctx, completionTime)
	dvvTriplet := types.DVVTriplet{
		DelegatorAccount:    red.DelegatorAccount,
		ValidatorSrcAccount: red.ValidatorSrcAccount,
		ValidatorDstAccount: red.ValidatorDstAccount}

	if len(timeSlice) == 0 {
		k.SetRedelegationQueueTimeSlice(ctx, completionTime, []types.DVVTriplet{dvvTriplet})
	} else {
		timeSlice = append(timeSlice, dvvTriplet)
		k.SetRedelegationQueueTimeSlice(ctx, completionTime, timeSlice)
	}
}

// Returns all the redelegation queue timeslices from time 0 until endTime
func (k Keeper) RedelegationQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := store.NewStore(ctx, k.storeKey)
	return store.Iterator(types.RedelegationQueueKey, sdk.InclusiveEndBytes(types.GetRedelegationTimeKey(endTime)))
}

// Returns a concatenated list of all the timeslices inclusively previous to
// currTime, and deletes the timeslices from the queue
func (k Keeper) DequeueAllMatureRedelegationQueue(ctx sdk.Context, currTime time.Time) (matureRedelegations []types.DVVTriplet) {
	store := store.NewStore(ctx, k.storeKey)

	// gets an iterator for all timeslices from time 0 until the current Blockheader time
	redelegationTimesliceIterator := k.RedelegationQueueIterator(ctx, ctx.BlockHeader().Time)
	for ; redelegationTimesliceIterator.Valid(); redelegationTimesliceIterator.Next() {
		timeslice := types.DVVTriplets{}
		value := redelegationTimesliceIterator.Value()
		k.cdc.MustUnmarshalBinaryBare(value, &timeslice)

		matureRedelegations = append(matureRedelegations, timeslice.Triplets...)
		store.Delete(redelegationTimesliceIterator.Key())
	}

	return matureRedelegations
}

// Perform a delegation, set/update everything necessary within the store.
// tokenSrc indicates the bond status of the incoming funds.
func (k Keeper) Delegate(
	ctx sdk.Context, delAddr AccountID, bondAmt sdk.Int, tokenSrc stakingexport.BondStatus,
	validator types.Validator, subtractAccount bool,
) (newShares sdk.Dec, err error) {

	// In some situations, the exchange rate becomes invalid, e.g. if
	// Validator loses all tokens due to slashing. In this case,
	// make all future delegations invalid.
	if validator.InvalidExRate() {
		return sdk.ZeroDec(), types.ErrDelegatorShareExRateInvalid
	}

	// Get or create the delegation object
	delegation, found := k.GetDelegation(ctx, delAddr, validator.OperatorAccount)
	if !found {
		delegation = types.NewDelegation(delAddr, validator.OperatorAccount, sdk.ZeroDec())
	}
	// call the appropriate hook if present
	if found {
		k.BeforeDelegationSharesModified(ctx, delAddr, validator.OperatorAccount)
	} else {
		k.BeforeDelegationCreated(ctx, delAddr, validator.OperatorAccount)
	}

	// if subtractAccount is true then we are
	// performing a delegation and not a redelegation, thus the source tokens are
	// all non bonded
	if subtractAccount {
		if tokenSrc == stakingexport.Bonded {
			panic("delegation token source cannot be bonded")
		}

		var sendName string
		switch {
		case validator.IsBonded():
			sendName = types.BondedPoolName
		case validator.IsUnbonding(), validator.IsUnbonded():
			sendName = types.NotBondedPoolName
		default:
			panic("invalid validator status")
		}

		coins := NewCoins(NewCoin(k.BondDenom(ctx), bondAmt))
		err := k.supplyKeeper.DelegateCoinsFromAccountToModule(ctx, sendName, coins)
		if err != nil {
			return sdk.Dec{}, err
		}
	} else {

		// potentially transfer tokens between pools, if
		switch {
		case tokenSrc == stakingexport.Bonded && validator.IsBonded():
			// do nothing
		case (tokenSrc == stakingexport.Unbonded || tokenSrc == stakingexport.Unbonding) && !validator.IsBonded():
			// do nothing
		case (tokenSrc == stakingexport.Unbonded || tokenSrc == stakingexport.Unbonding) && validator.IsBonded():
			// transfer pools
			k.notBondedTokensToBonded(ctx, bondAmt)
		case tokenSrc == stakingexport.Bonded && !validator.IsBonded():
			// transfer pools
			k.bondedTokensToNotBonded(ctx, bondAmt)
		default:
			panic("unknown token source bond status")
		}
	}

	validator, newShares = k.AddValidatorTokensAndShares(ctx, validator, bondAmt)

	// Update delegation
	delegation.Shares = delegation.Shares.Add(newShares)
	k.SetDelegation(ctx, delegation)
	// Call the after-modification hook
	k.AfterDelegationModified(ctx, delegation.DelegatorAccount, delegation.ValidatorAccount)

	return newShares, nil
}

// unbond a particular delegation and perform associated store operations
func (k Keeper) Unbond(
	ctx sdk.Context, delAddr AccountID, valAddr AccountID, shares sdk.Dec,
) (amount sdk.Int, err error) {

	// check if a delegation object exists in the store
	delegation, found := k.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return amount, types.ErrNoDelegatorForAddress
	}

	// call the before-delegation-modified hook
	k.BeforeDelegationSharesModified(ctx, delAddr, valAddr)

	// ensure that we have enough shares to remove
	if delegation.Shares.LT(shares) {
		return amount, sdkerrors.Wrap(types.ErrNotEnoughDelegationShares, delegation.Shares.String())
	}

	// get validator
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return amount, types.ErrNoValidatorFound
	}

	// subtract shares from delegation
	delegation.Shares = delegation.Shares.Sub(shares)

	isValidatorOperator := delegation.DelegatorAccount.Eq(validator.OperatorAccount)

	// if the delegation is the operator of the validator and undelegating will decrease the validator's self delegation below their minimum
	// trigger a jail validator
	if isValidatorOperator && !validator.Jailed &&
		validator.TokensFromShares(delegation.Shares).TruncateInt().LT(validator.MinSelfDelegation) {

		k.jailValidator(ctx, validator)
		validator = k.mustGetValidator(ctx, validator.OperatorAccount)
	}

	// remove the delegation
	if delegation.Shares.IsZero() {
		k.RemoveDelegation(ctx, delegation)
	} else {
		k.SetDelegation(ctx, delegation)
		// call the after delegation modification hook
		k.AfterDelegationModified(ctx, delegation.DelegatorAccount, delegation.ValidatorAccount)
	}

	// remove the shares and coins from the validator
	// NOTE that the amount is later (in keeper.Delegation) moved between staking module pools
	validator, amount = k.RemoveValidatorTokensAndShares(ctx, validator, shares)

	if validator.DelegatorShares.IsZero() && validator.IsUnbonded() {
		// if not unbonded, we must instead remove validator in EndBlocker once it finishes its unbonding period
		k.RemoveValidator(ctx, validator.OperatorAccount)
	}

	return amount, nil
}

// getBeginInfo returns the completion time and height of a redelegation, along
// with a boolean signaling if the redelegation is complete based on the source
// validator.
func (k Keeper) getBeginInfo(
	ctx sdk.Context, valSrcAddr AccountID,
) (completionTime time.Time, height int64, completeNow bool) {

	validator, found := k.GetValidator(ctx, valSrcAddr)

	// TODO: When would the validator not be found?
	switch {
	case !found || validator.IsBonded():

		// the longest wait - just unbonding period from now
		completionTime = ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
		height = ctx.BlockHeight()
		return completionTime, height, false

	case validator.IsUnbonded():
		return completionTime, height, true

	case validator.IsUnbonding():
		return validator.UnbondingTime, validator.UnbondingHeight, false

	default:
		panic(fmt.Sprintf("unknown validator status: %s", validator.Status))
	}
}

// Undelegate unbonds an amount of delegator shares from a given validator. It
// will verify that the unbonding entries between the delegator and validator
// are not exceeded and unbond the staked tokens (based on shares) by creating
// an unbonding object and inserting it into the unbonding queue which will be
// processed during the staking EndBlocker.
func (k Keeper) Undelegate(
	ctx sdk.Context, delAddr AccountID, valAddr AccountID, sharesAmount sdk.Dec,
) (time.Time, error) {

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return time.Time{}, types.ErrNoDelegatorForAddress
	}

	if k.HasMaxUnbondingDelegationEntries(ctx, delAddr, valAddr) {
		return time.Time{}, types.ErrMaxUnbondingDelegationEntries
	}

	returnAmount, err := k.Unbond(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return time.Time{}, err
	}

	// transfer the validator tokens to the not bonded pool
	if validator.IsBonded() {
		k.bondedTokensToNotBonded(ctx, returnAmount)
	}

	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
	ubd := k.SetUnbondingDelegationEntry(ctx, delAddr, valAddr, ctx.BlockHeight(), completionTime, returnAmount)
	k.InsertUBDQueue(ctx, ubd, completionTime)

	return completionTime, nil
}

// CompleteUnbonding completes the unbonding of all mature entries in the
// retrieved unbonding delegation object and returns the total unbonding balance
// or an error upon failure.
func (k Keeper) CompleteUnbonding(ctx sdk.Context, delAddr AccountID, valAddr AccountID) (Coins, error) {
	ubd, found := k.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	bondDenom := k.GetParams(ctx).BondDenom
	balances := NewCoins()
	ctxTime := ctx.BlockHeader().Time

	// loop through all the entries and complete unbonding mature entries
	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		if entry.IsMature(ctxTime) {
			ubd.RemoveEntry(int64(i))
			i--

			// track undelegation only when remaining or truncated shares are non-zero
			if !entry.Balance.IsZero() {
				amt := NewCoin(bondDenom, entry.Balance)
				err := k.supplyKeeper.UndelegateCoinsFromModuleToAccount(
					ctx, types.NotBondedPoolName, ubd.DelegatorAccount, NewCoins(amt),
				)
				if err != nil {
					return nil, err
				}

				balances = balances.Add(amt)
			}
		}
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		k.SetUnbondingDelegation(ctx, ubd)
	}

	return balances, nil
}

// begin unbonding / redelegation; create a redelegation record
func (k Keeper) BeginRedelegation(
	ctx sdk.Context, delAddr AccountID, valSrcAddr, valDstAddr AccountID, sharesAmount sdk.Dec,
) (completionTime time.Time, err error) {

	if bytes.Equal(valSrcAddr.StoreKey(), valDstAddr.StoreKey()) {
		return time.Time{}, types.ErrSelfRedelegation
	}

	dstValidator, found := k.GetValidator(ctx, valDstAddr)
	if !found {
		return time.Time{}, types.ErrBadRedelegationDst
	}

	srcValidator, found := k.GetValidator(ctx, valSrcAddr)
	if !found {
		return time.Time{}, types.ErrBadRedelegationDst
	}

	// check if this is a transitive redelegation
	if k.HasReceivingRedelegation(ctx, delAddr, valSrcAddr) {
		return time.Time{}, types.ErrTransitiveRedelegation
	}

	if k.HasMaxRedelegationEntries(ctx, delAddr, valSrcAddr, valDstAddr) {
		return time.Time{}, types.ErrMaxRedelegationEntries
	}

	returnAmount, err := k.Unbond(ctx, delAddr, valSrcAddr, sharesAmount)
	if err != nil {
		return time.Time{}, err
	}

	if returnAmount.IsZero() {
		return time.Time{}, types.ErrTinyRedelegationAmount
	}

	sharesCreated, err := k.Delegate(ctx, delAddr, returnAmount, srcValidator.GetStatus(), dstValidator, false)
	if err != nil {
		return time.Time{}, err
	}

	// create the unbonding delegation
	completionTime, height, completeNow := k.getBeginInfo(ctx, valSrcAddr)

	if completeNow { // no need to create the redelegation object
		return completionTime, nil
	}

	red := k.SetRedelegationEntry(
		ctx, delAddr, valSrcAddr, valDstAddr,
		height, completionTime, returnAmount, sharesAmount, sharesCreated,
	)
	k.InsertRedelegationQueue(ctx, red, completionTime)
	return completionTime, nil
}

// CompleteRedelegation completes the redelegations of all mature entries in the
// retrieved redelegation object and returns the total redelegation (initial)
// balance or an error upon failure.
func (k Keeper) CompleteRedelegation(
	ctx sdk.Context, delAddr AccountID, valSrcAddr, valDstAddr AccountID,
) (Coins, error) {

	red, found := k.GetRedelegation(ctx, delAddr, valSrcAddr, valDstAddr)
	if !found {
		return nil, types.ErrNoRedelegation
	}

	bondDenom := k.GetParams(ctx).BondDenom
	balances := NewCoins()
	ctxTime := ctx.BlockHeader().Time

	// loop through all the entries and complete mature redelegation entries
	for i := 0; i < len(red.Entries); i++ {
		entry := red.Entries[i]
		if entry.IsMature(ctxTime) {
			red.RemoveEntry(int64(i))
			i--

			if !entry.InitialBalance.IsZero() {
				balances = balances.Add(NewCoin(bondDenom, entry.InitialBalance))
			}
		}
	}

	// set the redelegation or remove it if there are no more entries
	if len(red.Entries) == 0 {
		k.RemoveRedelegation(ctx, red)
	} else {
		k.SetRedelegation(ctx, red)
	}

	return balances, nil
}

// ValidateUnbondAmount validates that a given unbond or redelegation amount is
// valied based on upon the converted shares. If the amount is valid, the total
// amount of respective shares is returned, otherwise an error is returned.
func (k Keeper) ValidateUnbondAmount(
	ctx sdk.Context, delAddr AccountID, valAddr AccountID, amt sdk.Int,
) (shares sdk.Dec, err error) {

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return shares, types.ErrNoValidatorFound
	}

	del, found := k.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return shares, types.ErrNoDelegation
	}

	shares, err = validator.SharesFromTokens(amt)
	if err != nil {
		return shares, err
	}

	sharesTruncated, err := validator.SharesFromTokensTruncated(amt)
	if err != nil {
		return shares, err
	}

	delShares := del.GetShares()
	if sharesTruncated.GT(delShares) {
		return shares, types.ErrBadSharesAmount
	}

	// Cap the shares at the delegation's shares. Shares being greater could occur
	// due to rounding, however we don't want to truncate the shares or take the
	// minimum because we want to allow for the full withdraw of shares from a
	// delegation.
	if shares.GT(delShares) {
		shares = delShares
	}

	return shares, nil
}