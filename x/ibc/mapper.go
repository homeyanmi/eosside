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

// Stores the sequence number of incoming IBC packet under "ingress/index".
func IngressSequenceKey() []byte {
	return []byte(fmt.Sprintf("ingress"))
}

// TODO add description
func (ibcm Mapper) GetIngressSequence(ctx sdk.Context) int64 {
	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey()

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
func (ibcm Mapper) SetIngressSequence(ctx sdk.Context, sequence int64) {
	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey()

	bz := marshalBinaryPanic(ibcm.cdc, sequence)
	store.Set(key, bz)
}

