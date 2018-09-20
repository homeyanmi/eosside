package ibc

import (
	"reflect"
	"os"
	"bytes"
	"fmt"

	"github.com/tendermint/tendermint/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	tmsdk "github.com/tendermint/tendermint/types"
	"github.com/tendermint/iavl"
)

func NewHandler(ibcm Mapper, ck bank.Keeper) sdk.Handler {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger.Info("IBC msg handler : ", "string", "xxx")
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case IBCRelayMsg:
			logger.Info("IBC msg handler : ", "type", "IBCRelayMsg")
			return handleIBCRelayMsg(ctx, ibcm, ck, msg)
		default:
			logger.Info("IBC msg handler : ", "type", "default")
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleIBCRelayMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg IBCRelayMsg) sdk.Result {
	
	//
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger.Info("IBC msg type : ", "string", "IBCRelayMsg")
	logger.Info("relay msg type : ", "string", msg.PayloadType)

	//
	switch msg.MsgType {
	case TRANSFER:
	    var transfer *Transfer
		err := ibcm.cdc.UnmarshalBinary(msg.Payload, &transfer)
		if err != nil {
			panic(err)
		}
		return handleTransfer(ctx, ibcm, ck, *transfer)
	default:
		return sdk.ErrUnknownRequest("unknown ibc relay msg!").Result()
	}
}

// IBCReceiveMsg adds coins to the destination address and creates an ingress IBC packet.
func handleTransfer(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg Transfer) sdk.Result {

	_, _, err := ck.AddCoins(ctx, msg.DestAddr, msg.Coins)
	if err != nil {
		return err.Result()
	}

	//
	return sdk.Result{}
}




