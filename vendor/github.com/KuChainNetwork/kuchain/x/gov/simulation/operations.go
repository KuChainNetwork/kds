package simulation

import (
	"math"
	"math/rand"
	"time"

	"github.com/KuChainNetwork/kuchain/chain/transaction/helpers"
	chainTypes "github.com/KuChainNetwork/kuchain/chain/types"
	kuSim "github.com/KuChainNetwork/kuchain/test/simulation"
	"github.com/KuChainNetwork/kuchain/x/gov/keeper"
	"github.com/KuChainNetwork/kuchain/x/gov/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

var initialProposalID = uint64(100000000000000)

// Simulation operation weights constants
const (
	OpWeightMsgDeposit = "op_weight_msg_deposit"
	OpWeightMsgVote    = "op_weight_msg_vote"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper, wContents []simulation.WeightedProposalContent,
) simulation.WeightedOperations {

	var (
		weightMsgDeposit int
		weightMsgVote    int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDeposit, &weightMsgDeposit, nil,
		func(_ *rand.Rand) {
			weightMsgDeposit = simappparams.DefaultWeightMsgDeposit
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgVote, &weightMsgVote, nil,
		func(_ *rand.Rand) {
			weightMsgVote = simappparams.DefaultWeightMsgVote
		},
	)

	// generate the weighted operations for the proposal contents
	var wProposalOps simulation.WeightedOperations

	for _, wContent := range wContents {
		wContent := wContent // pin variable
		var weight int
		appParams.GetOrGenerate(cdc, wContent.AppParamsKey, &weight, nil,
			func(_ *rand.Rand) { weight = wContent.DefaultWeight })

		wProposalOps = append(
			wProposalOps,
			simulation.NewWeightedOperation(
				weight,
				SimulateSubmitProposal(ak, bk, k, wContent.ContentSimulatorFn),
			),
		)
	}

	wGovOps := simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgDeposit,
			SimulateMsgDeposit(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgVote,
			SimulateMsgVote(ak, bk, k),
		),
	}

	return append(wProposalOps, wGovOps...)
}

// SimulateSubmitProposal simulates creating a msg Submit Proposal
// voting on the proposal, and subsequently slashing the proposal. It is implemented using
// future operations.
func SimulateSubmitProposal(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, contentSim simulation.ContentSimulatorFn,
) simulation.Operation {
	// The states are:
	// column 1: All validators vote
	// column 2: 90% vote
	// column 3: 75% vote
	// column 4: 40% vote
	// column 5: 15% vote
	// column 6: noone votes
	// All columns sum to 100 for simplicity, values chosen by @valardragon semi-arbitrarily,
	// feel free to change.
	numVotesTransitionMatrix, _ := simulation.CreateTransitionMatrix([][]int{
		{20, 10, 0, 0, 0, 0},
		{55, 50, 20, 10, 0, 0},
		{25, 25, 30, 25, 30, 15},
		{0, 15, 30, 25, 30, 30},
		{0, 0, 20, 30, 30, 30},
		{0, 0, 0, 10, 10, 25},
	})

	statePercentageArray := []float64{1, .9, .75, .4, .15, 0}
	curNumVotesState := 1

	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// 1) submit proposal now
		content := contentSim(r, ctx, accs)
		if content == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, _ := simulation.RandomAcc(r, accs)
		id := chainTypes.NewAccountIDFromAccAdd(simAccount.Address)

		deposit, skip, err := randomDeposit(r, ctx, ak, bk, k, id)
		switch {
		case skip:
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case err != nil:
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewKuMsgSubmitProposal(simAccount.Address, content, deposit, id)
		account := ak.GetAccount(ctx, id)
		spendable := bk.SpendableCoins(ctx, id)

		var fees Coins
		coins, hasNeg := spendable.SafeSub(deposit)
		if !hasNeg {
			fees, err = kuSim.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, err
			}
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{0}, // TODO: sim support new seq []uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		opMsg := simulation.NewOperationMsg(msg, true, "")

		// get the submitted proposal ID
		proposalID, err := k.GetProposalID(ctx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// 2) Schedule operations for votes
		// 2.1) first pick a number of people to vote.
		curNumVotesState = numVotesTransitionMatrix.NextState(r, curNumVotesState)
		numVotes := int(math.Ceil(float64(len(accs)) * statePercentageArray[curNumVotesState]))

		// 2.2) select who votes and when
		whoVotes := r.Perm(len(accs))

		// didntVote := whoVotes[numVotes:]
		whoVotes = whoVotes[:numVotes]
		votingPeriod := k.GetVotingParams(ctx).VotingPeriod

		fops := make([]simulation.FutureOperation, numVotes+1)
		for i := 0; i < numVotes; i++ {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simulation.FutureOperation{
				BlockTime: whenVote,
				Op:        operationSimulateMsgVote(ak, bk, k, accs[whoVotes[i]], int64(proposalID)),
			}
		}

		return opMsg, fops, nil
	}
}

// SimulateMsgDeposit generates a MsgDeposit with random values.
func SimulateMsgDeposit(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		proposalID, ok := randomProposalID(r, k, ctx, types.StatusDepositPeriod)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		depositAccountID := chainTypes.NewAccountIDFromAccAdd(simAccount.Address)
		deposit, skip, err := randomDeposit(r, ctx, ak, bk, k, depositAccountID)
		switch {
		case skip:
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case err != nil:
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewKuMsgDeposit(simAccount.Address, depositAccountID, proposalID, deposit)
		account := ak.GetAccount(ctx, depositAccountID)
		spendable := bk.SpendableCoins(ctx, depositAccountID)

		var fees Coins
		coins, hasNeg := spendable.SafeSub(deposit)
		if !hasNeg {
			fees, err = kuSim.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, err
			}
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{0}, // TODO: sim support new seq []uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgVote generates a MsgVote with random values.
func SimulateMsgVote(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simulation.Operation {
	return operationSimulateMsgVote(ak, bk, k, simulation.Account{}, -1)
}

func operationSimulateMsgVote(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
	simAccount simulation.Account, proposalIDInt int64) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		if simAccount.Equals(simulation.Account{}) {
			simAccount, _ = simulation.RandomAcc(r, accs)
		}

		var proposalID uint64

		switch {
		case proposalIDInt < 0:
			var ok bool
			proposalID, ok = randomProposalID(r, k, ctx, types.StatusVotingPeriod)
			if !ok {
				return simulation.NoOpMsg(types.ModuleName), nil, nil
			}
		default:
			proposalID = uint64(proposalIDInt)
		}
		voteAccountID := chainTypes.NewAccountIDFromAccAdd(simAccount.Address)
		option := randomVotingOption(r)
		msg := types.NewKuMsgVote(simAccount.Address, voteAccountID, proposalID, option)

		account := ak.GetAccount(ctx, voteAccountID)
		spendable := bk.SpendableCoins(ctx, voteAccountID)

		fees, err := kuSim.RandomFees(r, ctx, spendable)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{0}, // TODO: sim support new seq []uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// Pick a random deposit with a random denomination with a
// deposit amount between (0, min(balance, minDepositAmount))
// This is to simulate multiple users depositing to get the
// proposal above the minimum deposit amount
func randomDeposit(r *rand.Rand, ctx sdk.Context,
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, id chainTypes.AccountID,
) (deposit Coins, skip bool, err error) {
	spendable := bk.SpendableCoins(ctx, id)

	if spendable.Empty() {
		return nil, true, nil // skip
	}

	minDeposit := k.GetDepositParams(ctx).MinDeposit
	denomIndex := r.Intn(len(minDeposit))
	denom := minDeposit[denomIndex].Denom

	depositCoins := spendable.AmountOf(denom)
	if depositCoins.IsZero() {
		return nil, true, nil
	}

	maxAmt := depositCoins
	if maxAmt.GT(minDeposit[denomIndex].Amount) {
		maxAmt = minDeposit[denomIndex].Amount
	}

	amount, err := simulation.RandPositiveInt(r, maxAmt)
	if err != nil {
		return nil, false, err
	}

	return Coins{chainTypes.NewCoin(denom, amount)}, false, nil
}

// Pick a random proposal ID between the initial proposal ID
// (defined in gov GenesisState) and the latest proposal ID
// that matches a given Status.
// It does not provide a default ID.
func randomProposalID(r *rand.Rand, k keeper.Keeper,
	ctx sdk.Context, status types.ProposalStatus) (proposalID uint64, found bool) {
	proposalID, _ = k.GetProposalID(ctx)

	switch {
	case proposalID > initialProposalID:
		// select a random ID between [initialProposalID, proposalID]
		proposalID = uint64(simulation.RandIntBetween(r, int(initialProposalID), int(proposalID)))

	default:
		// This is called on the first call to this funcion
		// in order to update the global variable
		initialProposalID = proposalID
	}

	proposal, ok := k.GetProposal(ctx, proposalID)
	if !ok || proposal.Status != status {
		return proposalID, false
	}

	return proposalID, true
}

// Pick a random voting option
func randomVotingOption(r *rand.Rand) types.VoteOption {
	switch r.Intn(4) {
	case 0:
		return types.OptionYes
	case 1:
		return types.OptionAbstain
	case 2:
		return types.OptionNo
	case 3:
		return types.OptionNoWithVeto
	default:
		panic("invalid vote option")
	}
}