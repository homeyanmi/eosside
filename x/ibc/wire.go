package ibc

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(IBCTransferMsg{}, "cosmos-sdk/IBCTransferMsg", nil)
	cdc.RegisterConcrete(IBCRegisterMsg{}, "cosmos-sdk/IBCRegisterMsg", nil)
	cdc.RegisterConcrete(IBCRelayMsg{}, "cosmos-sdk/IBCRelayMsg", nil)
	//cdc.RegisterConcrete(IBCPacketXXX{}, "cosmos-sdk/IBCPacketXXX", nil)
	//cdc.RegisterInterface((*Payload)(nil), nil)
	//cdc.RegisterConcrete(&RegisterPayload{}, "cosmos-sdk/RegisterPayload",nil)
	//cdc.RegisterConcrete(&TransferPayload{}, "cosmos-sdk/TransferPayload",nil)
}
