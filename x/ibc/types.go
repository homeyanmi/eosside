package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

var (
	msgCdc *wire.Codec
)

func init() {
	msgCdc = wire.NewCodec()
}

// ----------------------------------
// IBCRelayMsg

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// IBCTransferMsg defines how another module can send an IBCPacket.
type IBCRelayMsg struct {
	PayloadType    int64
	Payload        []byte
	Sequence       int64
	Relayer        sdk.AccAddress
}

const (
	TRANSFER   	= 1
)

type Transfer struct {
	SrcAddr     string
	DestAddr    sdk.AccAddress
	Coins       sdk.Coins
}

// nolint
func (msg IBCRelayMsg) Type() string {
	return "ibc"
}

// x/bank/tx.go MsgSend.GetSigners()
func (msg IBCRelayMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Relayer}
}

// get the sign bytes for ibc transfer message
func (msg IBCRelayMsg) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// validate ibc transfer message
func (msg IBCRelayMsg) ValidateBasic() sdk.Error {
	return nil
}

