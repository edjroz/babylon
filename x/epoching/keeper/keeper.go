package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/babylonchain/babylon/x/epoching/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		hooks      types.EpochingHooks
		paramstore paramtypes.Subspace
		bk         types.BankKeeper
		stk        types.StakingKeeper
		router     *baseapp.MsgServiceRouter
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bk types.BankKeeper,
	stk types.StakingKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		hooks:      nil,
		bk:         bk,
		stk:        stk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetHooks sets the validator hooks
func (k *Keeper) SetHooks(eh types.EpochingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = eh

	return k
}

// SetMsgServiceRouter sets the msgServiceRouter
func (k *Keeper) SetMsgServiceRouter(router *baseapp.MsgServiceRouter) *Keeper {
	k.router = router
	return k
}
