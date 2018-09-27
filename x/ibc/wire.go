package ibc

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(EosTransferMsg{}, "cosmos-sdk/EosTransferMsg", nil)
	cdc.RegisterConcrete(SideTransferMsg{}, "cosmos-sdk/SideTransferMsg", nil)
}
