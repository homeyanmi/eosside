package ibc

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	//abci "github.com/tendermint/tendermint/abci/types"
	tmsdk "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/libs/log"
)

// IBC Mapper
type Mapper struct {
	key       sdk.StoreKey
	cdc       *wire.Codec
	codespace sdk.CodespaceType
}

// XXX: The Mapper should not take a CoinKeeper. Rather have the CoinKeeper
// take an Mapper.
func NewMapper(cdc *wire.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Mapper {
	// XXX: How are these codecs supposed to work?
	return Mapper{
		key:       key,
		cdc:       cdc,
		codespace: codespace,
	}
}

// XXX: This is not the public API. This will change in MVP2 and will henceforth
// only be invoked from another module directly and not through a user
// transaction.
// TODO: Handle invalid IBC packets and return errors.
func (ibcm Mapper) PostIBCPacket(ctx sdk.Context, dest string, packet IBCPacket) sdk.Error {
	// write everything into the state
	store := ctx.KVStore(ibcm.key)
	
	//==========================================================
	//
	index := ibcm.GetRequestSequence(store, packet.DestChain)
	packet.Sequence = index
	
	//
	bz, err := ibcm.cdc.MarshalBinary(packet)
	if err != nil {
		panic(err)
	}
	
	store.Set(RequestKey(packet.DestChain, index), bz)
	
	//
	ibcm.SetRequestSequence(store, packet.DestChain, index + 1)
	
	//==========================================================
	//
	index = ibcm.GetEgressSequence(store, dest)
	store.Set(EgressKey(dest, index), bz)
	
	//
	ibcm.SetEgressSequence(store, dest, index + 1)
	
	//
	return nil
}


// XXX: In the future every module is able to register it's own handler for
// handling it's own IBC packets. The "ibc" handler will only route the packets
// to the appropriate callbacks.
// XXX: For now this handles all interactions with the CoinKeeper.
// XXX: This needs to do some authentication checking.
func (ibcm Mapper) ReceiveIBCPacket(ctx sdk.Context, packet IBCPacket) sdk.Error {
	return nil
}

func (ibcm Mapper) ReceiveAck(ctx sdk.Context, ack Ack) (request *IBCPacket, err sdk.Error) {
	//
	store := ctx.KVStore(ibcm.key)
	index := ibcm.GetRequestSequence(store, ack.DestChain)
	if ack.Sequence < 0 || ack.Sequence >= index {
		return nil, sdk.ErrUnknownRequest("Ack failed")
	}
	
	//
	bz := store.Get(RequestKey(ack.DestChain, ack.Sequence))
	if bz == nil {
		return nil, sdk.ErrUnknownRequest("Ack failed")
	}
	
	/*
	//==========================================================
	//
	index = ibcm.getResponseSequence(store, ack.DestChain)
	store.Set(RequestKey(ack.DestChain, index), bz)
	
	//
	SetRequestSequence(index + 1)
	*/
	
	var res *IBCPacket
	unmarshalBinaryPanic(ibcm.cdc, bz, &res)
	return res, nil
}

// --------------------------
// Functions for accessing the underlying KVStore.

func marshalBinaryPanic(cdc *wire.Codec, value interface{}) []byte {
	res, err := cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	return res
}

func unmarshalBinaryPanic(cdc *wire.Codec, bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinary(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// Stores the sequence number of incoming IBC packet under "ingress/index".
func IngressSequenceKey(chain string) []byte {
	return []byte(fmt.Sprintf("ingress/%s", chain))
}

// TODO add description
func (ibcm Mapper) GetIngressSequence(ctx sdk.Context, chain string) int64 {
	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey(chain)

	bz := store.Get(key)
	if bz == nil {
		zero := marshalBinaryPanic(ibcm.cdc, int64(0))
		store.Set(key, zero)
		return 0
	}

	var res int64
	unmarshalBinaryPanic(ibcm.cdc, bz, &res)
	return res
}

// TODO add description
func (ibcm Mapper) SetIngressSequence(ctx sdk.Context, chain string, sequence int64) {
	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey(chain)

	bz := marshalBinaryPanic(ibcm.cdc, sequence)
	store.Set(key, bz)
}


// Stores an outgoing IBC packet under "egress/chain_id/index".
func EgressKey(chain string, index int64) []byte {
	return []byte(fmt.Sprintf("egress/%s/%d", chain, index))
}

// Stores the number of outgoing IBC packets under "egress/index".
func EgressSequenceKey(chain string) []byte {
	return []byte(fmt.Sprintf("egress/%s", chain))
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm Mapper) GetEgressSequence(store sdk.KVStore, chain string) int64 {
	key := EgressSequenceKey(chain)
	
	bz := store.Get(key)
	if bz == nil {
		zero := marshalBinaryPanic(ibcm.cdc, int64(0))
		store.Set(key, zero)
		return 0
	}
	
	var res int64
	unmarshalBinaryPanic(ibcm.cdc, bz, &res)
	return res
}

func (ibcm Mapper) SetEgressSequence(store sdk.KVStore, chain string, sequence int64) {
	key := EgressSequenceKey(chain)

	bz := marshalBinaryPanic(ibcm.cdc, sequence)
	store.Set(key, bz)
}


func RequestSequenceKey(chain string) [] byte {
	return []byte(fmt.Sprintf("request/%s", chain))
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm Mapper) GetRequestSequence(store sdk.KVStore, chain string) int64 {
	key := RequestSequenceKey(chain)
	
	bz := store.Get(key)
	if bz == nil {
		zero := marshalBinaryPanic(ibcm.cdc, int64(0))
		store.Set(key, zero)
		return 0
	}
	
	var res int64
	unmarshalBinaryPanic(ibcm.cdc, bz, &res)
	return res
}

func (ibcm Mapper) SetRequestSequence(store sdk.KVStore, chain string, sequence int64) {
	key := RequestSequenceKey(chain)

	bz := marshalBinaryPanic(ibcm.cdc, sequence)
	store.Set(key, bz)
}

func RequestKey(chain string, index int64) []byte {
	return []byte(fmt.Sprintf("request/%s/%d", chain, index))
}


func ResponseSequenceKey(chain string) [] byte {
	return []byte(fmt.Sprintf("response/%s", chain))
}


// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm Mapper) GetResponseSequence(store sdk.KVStore, chain string) int64 {
	key := ResponseSequenceKey(chain)
	
	bz := store.Get(key)
	if bz == nil {
		zero := marshalBinaryPanic(ibcm.cdc, int64(0))
		store.Set(key, zero)
		return 0
	}
	
	var res int64
	unmarshalBinaryPanic(ibcm.cdc, bz, &res)
	return res
}

func (ibcm Mapper) SetResponseSequence(store sdk.KVStore, chain string, sequence int64) {
	key := ResponseSequenceKey(chain)

	bz := marshalBinaryPanic(ibcm.cdc, sequence)
	store.Set(key, bz)
}

func ResponseKey(chain string, index int64) []byte {
	return []byte(fmt.Sprintf("response/%s/%d", chain, index))
}

// InitGenesis - store genesis parameters
//func InitGenesis(ctx sdk.Context, ibcm Mapper, chainid string, validators []abci.Validator) {
func InitGenesis(ctx sdk.Context, ibcm Mapper, chainid string, validators []*tmsdk.Validator) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger.Info("InitGenesis : ", "string", "xxx")
	logger.Info("InitGenesis ChainID : ", "string", chainid)

	//
	state := IBCChainState{
		ChainID: chainid,
		Validators: validators,
		LastBlockHash: nil,
		LastBlockHeight: 0,
	}
	
	for _, val := range validators {
		pubKey := val.PubKey
		address := pubKey.Address()
		logger.Info("InitGenesis address : ", "string", address)
		logger.Info("InitGenesis address : ", "string", val.Address)
	}
	
	ibcm.storeState(ctx, state);
	
	//
	conns := Connections {
		Size: 0,
		Chains: make([]string, 16),
	}
	ibcm.storeConnections(ctx, conns)
	
	//
	route := RouteTable {
		Size: 0,
		RouteMap: make([]Route, 16),
	}
	route.RouteMap[0].Dest = "local"
	route.RouteMap[0].Next = state.ChainID
	route.Size = route.Size + 1
	ibcm.storeRoute(ctx, route)
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm Mapper) storeState(ctx sdk.Context, state IBCChainState) {
	store := ctx.KVStore(ibcm.key)
	key := []byte("state")
	res, err := ibcm.cdc.MarshalBinary(state)
	if err != nil {
		panic(err)
	}
	
	store.Set(key, res)
}

func (ibcm Mapper) getState(ctx sdk.Context) *IBCChainState {
	store := ctx.KVStore(ibcm.key)
	key := []byte("state")
	bz := store.Get(key)
	var state *IBCChainState
	err := ibcm.cdc.UnmarshalBinary(bz, &state)
	if err != nil {
		panic(err)
	}
	
	return state
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm Mapper) storeConnections(ctx sdk.Context, conns Connections) {
	store := ctx.KVStore(ibcm.key)
	key := []byte("connections")
	res, err := ibcm.cdc.MarshalBinary(conns)
	if err != nil {
		panic(err)
	}
	
	store.Set(key, res)
}

func (ibcm Mapper) getConnections(ctx sdk.Context) *Connections {
	store := ctx.KVStore(ibcm.key)
	key := []byte("connections")
	bz := store.Get(key)
	var conns *Connections
	err := ibcm.cdc.UnmarshalBinary(bz, &conns)
	if err != nil {
		panic(err)
	}
	
	return conns
}

func (ibcm Mapper) addConnection(ctx sdk.Context, chain string) {
	
	//
	conns := ibcm.getConnections(ctx)
	
	//
	exist := false
	for i, c := range conns.Chains {
		if i >= int(conns.Size) {
			break
		}
		
		if c == chain {
			exist = true
			break
		}
	}

	if exist == true {
		return
	}
	
	conns.Chains[conns.Size] = chain
	conns.Size = conns.Size + 1
	ibcm.storeConnections(ctx, *conns)
}

func (ibcm Mapper) existConnection(ctx sdk.Context, chain string) bool {
	conns := ibcm.getConnections(ctx)
	for i, _ := range conns.Chains {
		if i >= int(conns.Size) {
			return false
		}
		
		if conns.Chains[i] == chain {
			return true
		}
	}
	
	return false
}

func (ibcm Mapper) setHub(ctx sdk.Context, hub string) {
	route := ibcm.getRoute(ctx)
	route.Hub = hub
	ibcm.storeRoute(ctx, *route)	
}

func (ibcm Mapper) getHub(ctx sdk.Context) string {
	route := ibcm.getRoute(ctx)
	return route.Hub	
}

func (ibcm Mapper) isHub(ctx sdk.Context, hub string) bool {
	route := ibcm.getRoute(ctx)
	return route.Hub == hub
}



// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm Mapper) storeRoute(ctx sdk.Context, route RouteTable) {
	store := ctx.KVStore(ibcm.key)
	key := []byte("routetable")
	res, err := ibcm.cdc.MarshalBinary(route)
	if err != nil {
		panic(err)
	}
	
	store.Set(key, res)
}

func (ibcm Mapper) getRoute(ctx sdk.Context) *RouteTable {
	store := ctx.KVStore(ibcm.key)
	key := []byte("routetable")
	bz := store.Get(key)
	var route *RouteTable
	err := ibcm.cdc.UnmarshalBinary(bz, &route)
	if err != nil {
		panic(err)
	}
	
	return route
}

func (ibcm Mapper) addRoute(ctx sdk.Context, dest string, next string) {
	//
	route := ibcm.getRoute(ctx)
	
	//
	exist := false
	for i, r := range route.RouteMap {
		if i >= int(route.Size) {
			break;
		}
		if r.Dest == dest {
			route.RouteMap[i].Dest = dest
			route.RouteMap[i].Next = next
			exist = true
			break;
		}
	}
	
	if exist == false {
		route.RouteMap[route.Size].Dest = dest
		route.RouteMap[route.Size].Next = next
		route.Size = route.Size + 1
	}
	ibcm.storeRoute(ctx, *route)
}

func (ibcm Mapper) route(ctx sdk.Context, dest string) string {
	route := ibcm.getRoute(ctx)
	for i, r := range route.RouteMap {
		if i >= int(route.Size) {
			break;
		}
        if r.Dest == dest {
        	return r.Next
        }
    }
	
	return route.Hub
}

