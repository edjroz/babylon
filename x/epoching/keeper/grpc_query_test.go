package keeper_test

import (
	"math/rand"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

// FuzzParamsQuery fuzzes queryClient.Params
// 1. Generate random param
// 2. When EpochInterval is 0, ensure `Validate` returns an error
// 3. Randomly set the param via query and check if the param has been updated
func FuzzParamsQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		// params generated by fuzzer
		params := types.DefaultParams()
		epochInterval := datagen.RandomInt(20)
		params.EpochInterval = epochInterval

		// test the case of EpochInterval < 2
		// after that, change EpochInterval to a random value until >=2
		if epochInterval < 2 {
			// validation should not pass with EpochInterval < 2
			require.Error(t, params.Validate())
			params.EpochInterval = uint64(rand.Int())
		}

		helper := testepoching.NewHelper(t)
		ctx, keeper, queryClient := helper.Ctx, helper.EpochingKeeper, helper.QueryClient
		wctx := sdk.WrapSDKContext(ctx)
		// if setParamsFlag == 0, set params
		setParamsFlag := rand.Intn(2)
		if setParamsFlag == 0 {
			keeper.SetParams(ctx, params)
		}
		req := types.QueryParamsRequest{}
		resp, err := queryClient.Params(wctx, &req)
		require.NoError(t, err)
		// if setParamsFlag == 0, resp.Params should be changed, otherwise default
		if setParamsFlag == 0 {
			require.Equal(t, params, resp.Params)
		} else {
			require.Equal(t, types.DefaultParams(), resp.Params)
		}
	})
}

// FuzzCurrentEpoch fuzzes queryClient.CurrentEpoch
// 1. generate a random number of epochs to increment
// 2. query the current epoch and boundary
// 3. compare them with the correctly calculated ones
func FuzzCurrentEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		increment := datagen.RandomInt(100)

		helper := testepoching.NewHelper(t)
		ctx, keeper, queryClient := helper.Ctx, helper.EpochingKeeper, helper.QueryClient
		wctx := sdk.WrapSDKContext(ctx)

		epochInterval := keeper.GetParams(ctx).EpochInterval
		for i := uint64(0); i < increment; i++ {
			// this ensures that IncEpoch is invoked only at the first header of each epoch
			ctx = ctx.WithBlockHeader(*datagen.GenRandomTMHeader("chain-test", i*epochInterval+1))
			wctx = sdk.WrapSDKContext(ctx)
			keeper.IncEpoch(ctx)
		}
		req := types.QueryCurrentEpochRequest{}
		resp, err := queryClient.CurrentEpoch(wctx, &req)
		require.NoError(t, err)
		require.Equal(t, increment, resp.CurrentEpoch)
		require.Equal(t, increment*epochInterval, resp.EpochBoundary)
	})
}

func FuzzEpochsInfo(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		numEpochs := datagen.RandomInt(10) + 1
		limit := datagen.RandomInt(10) + 1

		helper := testepoching.NewHelper(t)
		ctx, keeper, queryClient := helper.Ctx, helper.EpochingKeeper, helper.QueryClient
		wctx := sdk.WrapSDKContext(ctx)

		// enque the first block of the numEpochs'th epoch
		epochInterval := keeper.GetParams(ctx).EpochInterval
		for i := uint64(0); i < numEpochs-1; i++ {
			for j := uint64(0); j < epochInterval; j++ {
				helper.GenAndApplyEmptyBlock()
			}
		}

		// get epoch msgs
		req := types.QueryEpochsInfoRequest{
			Pagination: &query.PageRequest{
				Limit: limit,
			},
		}
		resp, err := queryClient.EpochsInfo(wctx, &req)
		require.NoError(t, err)

		require.Equal(t, testepoching.Min(numEpochs, limit), uint64(len(resp.Epochs)))
		for i, epoch := range resp.Epochs {
			require.Equal(t, uint64(i), epoch.EpochNumber)
		}
	})
}

// FuzzEpochMsgsQuery fuzzes queryClient.EpochMsgs
// 1. randomly generate msgs and limit in pagination
// 2. check the returned msg was previously enqueued
// NOTE: Msgs in QueryEpochMsgsResponse are out-of-order
func FuzzEpochMsgsQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		numMsgs := uint64(rand.Int() % 100)
		limit := uint64(rand.Int()%100) + 1

		txidsMap := map[string]bool{}
		helper := testepoching.NewHelper(t)
		ctx, keeper, queryClient := helper.Ctx, helper.EpochingKeeper, helper.QueryClient
		wctx := sdk.WrapSDKContext(ctx)
		// enque a random number of msgs with random txids
		for i := uint64(0); i < numMsgs; i++ {
			txid := datagen.GenRandomByteArray(32)
			txidsMap[string(txid)] = true
			queuedMsg := types.QueuedMessage{
				TxId: txid,
				Msg:  &types.QueuedMessage_MsgDelegate{MsgDelegate: &stakingtypes.MsgDelegate{}},
			}
			keeper.EnqueueMsg(ctx, queuedMsg)
		}
		// get epoch msgs
		req := types.QueryEpochMsgsRequest{
			EpochNum: 0,
			Pagination: &query.PageRequest{
				Limit: limit,
			},
		}
		resp, err := queryClient.EpochMsgs(wctx, &req)
		require.NoError(t, err)

		require.Equal(t, testepoching.Min(uint64(len(txidsMap)), limit), uint64(len(resp.Msgs)))
		for idx := range resp.Msgs {
			_, ok := txidsMap[string(resp.Msgs[idx].TxId)]
			require.True(t, ok)
		}

		// epoch 1 is out of scope
		req = types.QueryEpochMsgsRequest{
			EpochNum: 1,
			Pagination: &query.PageRequest{
				Limit: limit,
			},
		}
		_, err = queryClient.EpochMsgs(wctx, &req)
		require.Error(t, err)
	})
}

// FuzzEpochMsgs fuzzes queryClient.EpochValSet
// TODO (stateful tests): create some random validators and check if the resulting validator set is consistent or not (require mocking MsgWrappedCreateValidator)
func FuzzEpochValSetQuery(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelperWithValSet(t)
		ctx, queryClient := helper.Ctx, helper.QueryClient

		limit := uint64(rand.Int() % 100)
		req := &types.QueryEpochValSetRequest{
			EpochNum: 0,
			Pagination: &query.PageRequest{
				Limit: limit,
			},
		}

		resp, err := queryClient.EpochValSet(ctx, req)
		require.NoError(t, err)

		// generate a random number of new blocks
		numIncBlocks := rand.Uint64()%1000 + 1
		for i := uint64(0); i < numIncBlocks; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}

		// check whether the validator set remains the same or not
		resp2, err := queryClient.EpochValSet(ctx, req)
		require.NoError(t, err)
		require.Equal(t, len(resp.Validators), len(resp2.Validators))
		for i := range resp2.Validators {
			require.Equal(t, resp.Validators[i].Addr, resp2.Validators[i].Addr)
		}
	})
}
