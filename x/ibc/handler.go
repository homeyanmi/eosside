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
		case IBCTransferMsg:
			logger.Info("IBC msg handler : ", "type", "IBCTransferMsg")
			return handleIBCTransferMsg(ctx, ibcm, ck, msg)
		case IBCRelayMsg:
			logger.Info("IBC msg handler : ", "type", "IBCRelayMsg")
			return handleIBCRelayMsg(ctx, ibcm, ck, msg)
		case IBCRegisterMsg:
			logger.Info("IBC msg handler : ", "type", "IBCRegisterMsg")
			return handleIBCRegisterMsg(ctx, ibcm, ck, msg)
		default:
			logger.Info("IBC msg handler : ", "type", "default")
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleIBCTransferMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg IBCTransferMsg) sdk.Result {

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger.Info("src chain : ", "id", msg.SrcChain)
	logger.Info("dest chain : ", "id", msg.DestChain)
	
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	if state.ChainID != msg.SrcChain {
		return sdk.ErrUnknownRequest("This Register msg should be ignored!").Result()
	}
	
	{
		_, _, err := ck.SubtractCoins(ctx, msg.SrcAddr, msg.Coins)
		if err != nil {
			return err.Result()
		}
	}
	
	//
    transfer := Transfer {
    	SrcAddr: msg.SrcAddr,
    	DestAddr: msg.DestAddr,
    	Coins: msg.Coins,
    }
    
	//
	bz, err := ibcm.cdc.MarshalBinary(transfer)
	if err != nil {
		return sdk.ErrUnknownRequest("Transfer marshal binary failed!").Result()
	}
	
	res := post(ctx, ibcm, msg.DestChain, TRANSFER, bz)
	if res.IsOK() == false {
		return res
	}

	return sdk.Result{}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleIBCRegisterMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg IBCRegisterMsg) sdk.Result {
	
	//
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger.Info("IBC msg type : ", "string", "IBCRegisterMsg")
	logger.Info("IBC msg SrcChain : ", "string", msg.SrcChain)
	logger.Info("IBC msg DestChain : ", "string", msg.DestChain)
	
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	if state.ChainID != msg.SrcChain {
		return sdk.ErrUnknownRequest("This Register msg should be ignored!").Result()
	}
	
	//
	sync := Sync {
		State: *state,
	}
	
	//
	bz, err := ibcm.cdc.MarshalBinary(sync)
	if err != nil {
		return sdk.ErrUnknownRequest("SYNC marshal binary failed!").Result()
	}
	
	// because this is the register message, there is no connection chain and hub,
	// so post this packet derectly to dest chain
	//
	res := post_direct(ctx, ibcm, msg.DestChain, SYNC, bz)
	if res.IsOK() == false {
		return res
	}
	
	// now we think connection is ok
	//
	//ibcm.addConnection(ctx, msg.DestChain)
	ibcm.setHub(ctx, msg.DestChain)

	return sdk.Result{}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleIBCRelayMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg IBCRelayMsg) sdk.Result {
	
	//
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger.Info("IBC msg type : ", "string", "IBCRelayMsg")
	logger.Info("relay msg type : ", "string", msg.MsgType)
	logger.Info("relay msg SrcChain : ", "string", msg.SrcChain)
	logger.Info("relay msg DestChain : ", "string", msg.DestChain)
	
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	if msg.DestChain != state.ChainID {
		return sdk.ErrUnknownRequest("This IBCRelay msg should be ignored!").Result()
	}
	
	//
	switch msg.MsgType {
	case COMMITUPDATE:
	    var commit *CommitUpdate
		err := ibcm.cdc.UnmarshalBinary(msg.Payload, &commit)
		if err != nil {
			panic(err)
		}
		return handleCommitUpdate(ctx, ibcm, ck, msg.SrcChain, msg.DestChain, *commit)
	case NEWIBCPACKET:
		var newpacket *NewIBCPacket
		err := ibcm.cdc.UnmarshalBinary(msg.Payload, &newpacket)
		if err != nil {
			panic(err)
		}
		return handleNewIBCPacket(ctx, ibcm, ck, msg.SrcChain, msg.DestChain, *newpacket)
	default:
		return sdk.ErrUnknownRequest("unknown ibc relay msg!").Result()
	}
}

// IBCTransferMsg deducts coins from the account and creates an egress IBC packet.
func handleCommitUpdate(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, from string, to string, msg CommitUpdate) sdk.Result {
	
	//
	store := ctx.KVStore(ibcm.key)
	res, err := ibcm.cdc.MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	store.Set([]byte(fmt.Sprintf("chaincommit/%s", from)), res)

	//
	return sdk.Result{}
}

func handleNewIBCPacket(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, from string, to string, msg NewIBCPacket) sdk.Result {
	//
	seq := ibcm.GetIngressSequence(ctx, from)
	if msg.Sequence != seq {
		return ErrInvalidSequence(ibcm.codespace).Result()
	}
	
	//
	var packet *IBCPacket
	err := ibcm.cdc.UnmarshalBinary(msg.Packet, &packet)
	if err != nil {
		panic(err)
	}
	
	// proof
	//
	if packet.PayloadType != SYNC && packet.PayloadType != ACK {
		key := []byte(fmt.Sprintf("egress/%s/%d", to, msg.Sequence))
		res := proof(ctx, ibcm, from, msg.Proof, key, msg.Packet)
		if res.IsOK() == false {
			return res
		}
	}
	
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	if state.ChainID != packet.DestChain {
		//
		forward(ctx, ibcm, ck, packet)
	} else {
		//
		switch packet.PayloadType {
		case SYNC:
			var sync *Sync
			err := ibcm.cdc.UnmarshalBinary(packet.Payload, &sync)
			if err != nil {
				panic(err)
			}
			
			//
			handleSYNC(ctx, ibcm, ck, packet.SrcChain, packet.DestChain, packet.Sequence, *sync)
		case TRANSFER:
			var transfer *Transfer
			err := ibcm.cdc.UnmarshalBinary(packet.Payload, &transfer)
			if err != nil {
				panic(err)
			}
			handleTransfer(ctx, ibcm, ck, packet.SrcChain, packet.DestChain, packet.Sequence, *transfer)
		case ACK:
			var ack *Ack
			err := ibcm.cdc.UnmarshalBinary(packet.Payload, &ack)
			if err != nil {
				panic(err)
			}
			handleAck(ctx, ibcm, ck, packet.SrcChain, packet.DestChain, packet.Sequence, *ack)
		default:
			return sdk.ErrUnknownRequest("IBC packet type is unknown!").Result()
		}
	}
	
	//
	ibcm.SetIngressSequence(ctx, from, seq+1)
	return sdk.Result{}
}

func handleSYNC(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, from string, to string, sequence int64, msg Sync) sdk.Result {
	//
	store := ctx.KVStore(ibcm.key)
	bz, err := ibcm.cdc.MarshalBinary(msg.State)
	if err != nil {
		panic(err)
	}
	store.Set([]byte(fmt.Sprintf("chainstate/%s", from)), bz)
	
	//
	res := reply_direct(ctx, ibcm, from, sequence, OK)
	if res.IsOK() == false {
		return res
	}
	
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	//
	if ibcm.isHub(ctx, from) == true {
		return sdk.Result{}
	}
	
	//
	sync := Sync {
		State: *state,
	}
	
	//
	bz, err = ibcm.cdc.MarshalBinary(sync)
	if err != nil {
		return sdk.ErrUnknownRequest("SYNC marshal binary failed!").Result()
	}
	
	// because this is the register message, there is no connection chain and hub,
	// so post this packet derectly to dest chain
	//
	res = post_direct(ctx, ibcm, from, SYNC, bz)
	if res.IsOK() == false {
		return res
	}
	
	//
	//ibcm.addConnection(ctx, from)
	
	return sdk.Result{}
}

// IBCReceiveMsg adds coins to the destination address and creates an ingress IBC packet.
func handleTransfer(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, from string, to string, sequence int64, msg Transfer) sdk.Result {

	_, _, err := ck.AddCoins(ctx, msg.DestAddr, msg.Coins)
	if err != nil {
		return err.Result()
	}

	//
	return sdk.Result{}
}

// IBCReceiveMsg adds coins to the destination address and creates an ingress IBC packet.
func handleAck(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, from string, to string, sequence int64, msg Ack) sdk.Result {
	//
	request, _ := ibcm.ReceiveAck(ctx, msg)
	
	//
	if request.PayloadType == SYNC {
		ibcm.addConnection(ctx, request.DestChain)
		ibcm.addRoute(ctx, request.DestChain, request.DestChain)
	}
	
	if msg.Code == OK {
		
	}
	
	//
	return sdk.Result{}
}


func proof(ctx sdk.Context, ibcm Mapper, chain string, proof_data []byte, key []byte, data []byte) sdk.Result {
	//
	store := ctx.KVStore(ibcm.key)
	chain_state := store.Get([]byte(fmt.Sprintf("chainstate/%s", chain)))
	var state *IBCChainState
	err := ibcm.cdc.UnmarshalBinary(chain_state, &state)
	if err != nil {
		return sdk.ErrUnknownRequest("chain state is unknown!").Result()
	}
	
	//
	chain_commit := store.Get([]byte(fmt.Sprintf("chaincommit/%s", chain)))
	var commit *CommitUpdate
	err = ibcm.cdc.UnmarshalBinary(chain_commit, &commit)
	if err != nil {
		return sdk.ErrUnknownRequest("chain commit is unknown!").Result()
	}
	
	//
	chainID := state.ChainID
	vals := state.Validators
	valSet := tmsdk.NewValidatorSet(vals)
	
	var blockID tmsdk.BlockID
	for _, pc := range commit.Commit.Precommits {
		if pc != nil {
			blockID = pc.BlockID
		}
	}
	if blockID.IsZero() {
		return sdk.ErrUnknownRequest("Commit is unknown").Result()
	}
	
	err = valSet.VerifyCommit(chainID, blockID, commit.Header.Height, &commit.Commit)
	if err != nil {
		return sdk.ErrUnknownRequest("Commit is verify failed!").Result()
	}
	
	//
	var app_proof *sdk.AppProof
	err = ibcm.cdc.UnmarshalBinary(proof_data, &app_proof)
	if err != nil {
		return sdk.ErrUnknownRequest("IBC Packet proof is unknown!").Result()
	}
	var store_proof *iavl.RangeProof
	err = ibcm.cdc.UnmarshalBinary(app_proof.StoreProof, &store_proof)
	if err != nil {
		return sdk.ErrUnknownRequest("Commit is verify failed!").Result()
	}
	var iavl_proof *sdk.CommitInfo
	err = ibcm.cdc.UnmarshalBinary(app_proof.IavlProof, &iavl_proof)
	if err != nil {
		return sdk.ErrUnknownRequest("Commit is verify failed!").Result()
	}
	
	//
	var app_store_root_hash []byte
	app_store_name := "ibc"
	for _, store_info := range iavl_proof.StoreInfos {
		//logger.Info("app state tree name : ", "string", store_info.Name)
		//logger.Info("app state tree version : ", "string", store_info.Core.CommitID.Version)
		//logger.Info("app state tree hash : ", "string", hex.EncodeToString(store_info.Core.CommitID.Hash))
		if store_info.Name == app_store_name {
			app_store_root_hash = store_info.Core.CommitID.Hash
		}
	}
	
	
	err = store_proof.Verify(app_store_root_hash)
	
	//logger.Info("abci query proof tree", "proof", store_proof.String())
	//logger.Info("abci query key", "key", app_res.Key)
	//logger.Info("abci query Value", "value", app_res.Value)
	//logger.Info("abci query Height", "height", app_res.Height)
	//logger.Info("abci query proof tree", "hash", hex.EncodeToString(store_proof.ComputeRootHash()))
	//logger.Info("header app hash", "hash", commit_res.Header.AppHash)
	//logger.Info("header height", "height", commit_res.Header.Height)
	//logger.Info("app state root", "string", hex.EncodeToString(iavl_proof.Hash()))
	
	
	if err != nil {
		return sdk.ErrUnknownRequest("Didn't validate store!").Result()
	}
	
	if bytes.Equal(iavl_proof.Hash(), commit.Header.AppHash) == false {
		return sdk.ErrUnknownRequest("Didn't validate app hash!").Result()
	}
	
	//
	err = store_proof.VerifyItem(key, data)
	if err != nil {
		return sdk.ErrUnknownRequest("Didn't validate key & value!").Result()
	}
	
	//
	return sdk.Result{}
}


func next(ctx sdk.Context, ibcm Mapper, dest string) string {
	return ibcm.route(ctx, dest)
}

func forward(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg *IBCPacket) sdk.Result {
	
	//
	nextChain := next(ctx, ibcm, msg.DestChain)
	if nextChain == "" {
		return sdk.ErrUnknownRequest("unreachable!").Result()
	}
	
	err := ibcm.PostIBCPacket(ctx, nextChain, *msg)
	if err != nil {
		return sdk.ErrUnknownRequest("Relay IBC Packet failed").Result()
	}
	
	return sdk.Result{}
}

func post(ctx sdk.Context, ibcm Mapper, dest string, msgtype int64, msgdata []byte) sdk.Result {
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	//
	packet := CreateIBCPacket(state.ChainID, dest, msgtype, msgdata)
	
	//
	nextChain := next(ctx, ibcm, dest)
	if nextChain == "" {
		return sdk.ErrUnknownRequest("unreachable!").Result()
	}
	
	//
	err := ibcm.PostIBCPacket(ctx, nextChain, packet)
	if err != nil {
		return sdk.ErrUnknownRequest("IBC packet post failed").Result()
	}
	
	//
	

	return sdk.Result{}
}

func post_direct(ctx sdk.Context, ibcm Mapper, dest string, msgtype int64, msgdata []byte) sdk.Result {
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	//
	packet := CreateIBCPacket(state.ChainID, dest, msgtype, msgdata)
	
	//
	err := ibcm.PostIBCPacket(ctx, dest, packet)
	if err != nil {
		return sdk.ErrUnknownRequest("IBC packet post failed").Result()
	}

	return sdk.Result{}
}

func reply(ctx sdk.Context, ibcm Mapper, from string, sequence int64, code int64) sdk.Result {
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	//
	ack := Ack {
		Code: code,
		SrcChain: from,
		DestChain: state.ChainID,
		Sequence: sequence,
	}
	
	//
	bz, err := ibcm.cdc.MarshalBinary(ack)
	if err != nil {
		return sdk.ErrUnknownRequest("ACK marshal binary failed!").Result()
	}
	
	//
	packet := CreateIBCPacket(state.ChainID, from, ACK, bz)
	
	//
	nextChain := next(ctx, ibcm, from)
	if nextChain == "" {
		return sdk.ErrUnknownRequest("unreachable!").Result()
	}
	
	//
	{
		err := ibcm.PostIBCPacket(ctx, nextChain, packet)
		if err != nil {
			return sdk.ErrUnknownRequest("IBC packet post failed").Result()
		}
	}

	return sdk.Result{}
}

func reply_direct(ctx sdk.Context, ibcm Mapper, from string, sequence int64, code int64) sdk.Result {
	//
	state := ibcm.getState(ctx)
	if state == nil {
		return sdk.ErrUnknownRequest("There is no state in this chain!").Result()
	}
	
	//
	ack := Ack {
		Code: code,
		SrcChain: from,
		DestChain: state.ChainID,
		Sequence: sequence,
	}
	
	//
	bz, err := ibcm.cdc.MarshalBinary(ack)
	if err != nil {
		return sdk.ErrUnknownRequest("ACK marshal binary failed!").Result()
	}
	
	//
	packet := CreateIBCPacket(state.ChainID, from, ACK, bz)
	
	//
	{
		err := ibcm.PostIBCPacket(ctx, from, packet)
		if err != nil {
			return sdk.ErrUnknownRequest("IBC packet post failed").Result()
		}
	}

	return sdk.Result{}
}




