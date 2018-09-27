package ibc

import (
	"reflect"
	"os"

	"github.com/tendermint/tendermint/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

func NewHandler(ibcm Mapper, ck bank.Keeper) sdk.Handler {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger.Info("IBC msg handler : ", "string", "xxx")
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case SideTransferMsg:
			logger.Info("IBC msg handler : ", "type", "TransferMsg")
			return handleSideTransferMsg(ctx, ibcm, ck, msg)	
		case EosTransferMsg:
			logger.Info("IBC msg handler : ", "type", "EosTransferMsg")
			return handleEosTransferMsg(ctx, ibcm, ck, msg)
		default:
			logger.Info("IBC msg handler : ", "type", "default")
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleEosTransferMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg EosTransferMsg) sdk.Result {
	//
	seq := ibcm.GetIngressSequence(ctx)
	if msg.Sequence != seq {
		return ErrInvalidSequence(ibcm.codespace).Result()
	}
	
	//
	_, _, err := ck.AddCoins(ctx, msg.DestAddr, msg.Coins)
	if err != nil {
		return err.Result()
	}
	
	//
	ibcm.SetIngressSequence(ctx, seq+1)
	
	//
	return sdk.Result{}
}

// handleTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleSideTransferMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg SideTransferMsg) sdk.Result {
	//
	{
		_, _, err := ck.SubtractCoins(ctx, msg.SrcAddr, msg.Coins)
		if err != nil {
			return err.Result()
		}
	}
    
	//
	bz, err := ibcm.cdc.MarshalBinary(msg)
	if err != nil {
		return sdk.ErrUnknownRequest("Transfer marshal binary failed!").Result()
	}
	
	// write everything into the state
	store := ctx.KVStore(ibcm.key)

	//
	index := ibcm.GetEgressSequence(store)
	store.Set(EgressKey(index), bz)
	
	//
	ibcm.SetEgressSequence(store, index + 1)
	
	//
	return sdk.Result{}
}
