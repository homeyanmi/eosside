package app

import (
	"os"
	"time"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	eos "github.com/eoscanada/eos-go"
	token "github.com/eoscanada/eos-go/token"
	"github.com/blockchain-develop/eosside/x/ibc"
)

const (
	flagEosAccount   = "eos-account"
)

// IBC relay command
func SideTransferCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "transfer",
		Run: func(cmd *cobar.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// build the message
			msg, err := buildSideTransferMsg(from)
			if err != nil {
				return err
			}

			// get password
			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
			return nil
		},
	}
	
	cmd.Flags().String(flagEosAccount, "", "eos account to transfer")
	cmd.MarkFlagRequired(flagEosAccount)
	viper.BindPFlag(flagEosAccount, cmd.Flags().Lookup(flagEosAccount))

	return cmd
}

func buildSideTransferMsg(from sdk.AccAddress) (sdk.Msg, error) {
	//
	amount := viper.GetString(flagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return nil, err
	}
	
	//
	msg := ibc.SideTransferMsg {
		SrcChain: from,
		Coins: coins,
		DestAddr: viper.GetString(flagEosAccount),
	}

	return msg, nil
}
