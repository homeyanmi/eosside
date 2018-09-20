package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmsdk "github.com/tendermint/tendermint/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

var (
	msgCdc *wire.Codec
)

func init() {
	msgCdc = wire.NewCodec()
}

// ----------------------------------
// IBCTransferMsg

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// IBCTransferMsg defines how another module can send an IBCPacket.
type IBCTransferMsg struct {
	SrcAddr   sdk.AccAddress
	DestAddr  sdk.AccAddress
	Coins     sdk.Coins
	SrcChain  string
	DestChain string
}

// nolint
func (msg IBCTransferMsg) Type() string {
	return "ibc"
}

// x/bank/tx.go MsgSend.GetSigners()
func (msg IBCTransferMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.SrcAddr}
}

// get the sign bytes for ibc transfer message
func (msg IBCTransferMsg) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// validate ibc transfer message
func (msg IBCTransferMsg) ValidateBasic() sdk.Error {
	if msg.SrcChain == msg.DestChain {
		return ErrIdenticalChains(DefaultCodespace).TraceSDK("")
	}
	if !msg.Coins.IsValid() {
		return sdk.ErrInvalidCoins("")
	}
	return nil
}

// ----------------------------------
// IBCRegisterMsg

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// IBCTransferMsg defines how another module can send an IBCPacket.
type IBCRegisterMsg struct {
	SrcChain  string
	DestChain string
	FromAddr   sdk.AccAddress
}

// nolint
func (msg IBCRegisterMsg) Type() string {
	return "ibc"
}

// x/bank/tx.go MsgSend.GetSigners()
func (msg IBCRegisterMsg) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddr}
}

// get the sign bytes for ibc transfer message
func (msg IBCRegisterMsg) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// validate ibc transfer message
func (msg IBCRegisterMsg) ValidateBasic() sdk.Error {
	if msg.SrcChain == msg.DestChain {
		return ErrIdenticalChains(DefaultCodespace).TraceSDK("")
	}
	return nil
}

// ----------------------------------
// IBCRelayMsg

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// IBCTransferMsg defines how another module can send an IBCPacket.
type IBCRelayMsg struct {
	SrcChain   string
	DestChain  string
	MsgType    int64
	Payload    []byte
	Relayer    sdk.AccAddress
}

const (
	COMMITUPDATE   	= 1
	NEWIBCPACKET    = 2
)

type CommitUpdate struct {
	Header tmsdk.Header
	Commit tmsdk.Commit
}

type NewIBCPacket struct {
	Packet      []byte
	Sequence    int64
	Proof       []byte
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


// ------------------------------
// IBCPacket

// nolint - TODO rename to Packet as IBCPacket stutters (golint)
// IBCPacket defines a piece of data that can be send between two separate
// blockchains.
type IBCPacket struct {
	SrcChain  		string
	DestChain 		string
	Sequence        int64
	PayloadType     int64
	Payload         []byte
}

func CreateIBCPacket(srcChain string, destChain string, payloadType int64, payload []byte) IBCPacket {
	return IBCPacket{
		SrcChain:  srcChain,
		DestChain: destChain,
		Payload: payload,
		PayloadType: payloadType,
	}
}

const (
	SYNC    	= 1
	FIN         = 2
	TRANSFER   	= 3
	ACK         = 4
)

type Sync struct {
	State           IBCChainState
}

type IBCChainState struct {
	ChainID            string
	Validators         []*tmsdk.Validator
	LastBlockHash      []byte
	LastBlockHeight    int64
}

type Fin struct {
	
}

type Transfer struct {
	SrcAddr     sdk.AccAddress
	DestAddr    sdk.AccAddress
	Coins       sdk.Coins
}

//==========================================================
const (
	OK	    	= 1
	UNREACHABLE = 2
	TIMEOUT   	= 3
)

type Ack struct {
	Code      int64
	SrcChain  		string
	DestChain 		string
	Sequence        int64
	Payload   []byte
}




//--------------------------------------------------------------------
type Connections struct {
	Size       int64
	Chains     []string
}

type Route struct {
	Dest    string
	Next    string
}

type RouteTable struct {
	Hub         string
	Size        int64
	RouteMap    []Route
}

