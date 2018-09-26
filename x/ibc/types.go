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
// EosTransferMsg

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// IBCTransferMsg defines how another module can send an IBCPacket.
type EosTransferMsg struct {
	SrcAddr        string
	DestAddr       sdk.AccAddress
	Coins          sdk.Coins
	Sequence       int64
	Relayer        sdk.AccAddress
}

// nolint
func (msg EosTransferMsg) Type() string {
	return "ibc"
}

// x/bank/tx.go MsgSend.GetSigners()
func (msg EosTransferMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Relayer}
}

// get the sign bytes for ibc transfer message
func (msg EosTransferMsg) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// validate ibc transfer message
func (msg EosTransferMsg) ValidateBasic() sdk.Error {
	return nil
}



// ----------------------------------
// SideTransferMsg

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// IBCTransferMsg defines how another module can send an IBCPacket.
type SideTransferMsg struct {
	SrcAddr       sdk.AccAddress
	DestAddr      string
	Coins         sdk.Coins
}

// nolint
func (msg SideTransferMsg) Type() string {
	return "ibc"
}

// x/bank/tx.go MsgSend.GetSigners()
func (msg SideTransferMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.SrcAddr}
}

// get the sign bytes for ibc transfer message
func (msg SideTransferMsg) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// validate ibc transfer message
func (msg SideTransferMsg) ValidateBasic() sdk.Error {
	return nil
}

