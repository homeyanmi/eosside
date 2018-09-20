package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagDestChain  = "destchain"
	flagHubChain = "hubchain"
)

// IBC transfer command
func IBCRegisterCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "register",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// build the message
			msg, err := buildRegisterMsg(from)
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

	cmd.Flags().String(flagHubChain, "", "hub chain id to register")
	cmd.MarkFlagRequired(flagHubChain)
	viper.BindPFlag(flagHubChain, cmd.Flags().Lookup(flagHubChain))
	return cmd
}

func buildRegisterMsg(from sdk.AccAddress) (sdk.Msg, error) {
	
	//
	SrcChain := viper.GetString(client.FlagChainID)
	HubChain := viper.GetString(flagHubChain)
	
	//
	msg := ibc.IBCRegisterMsg {
		SrcChain: SrcChain,
		DestChain: HubChain,
		FromAddr: from,
	}

	return msg, nil
}

// IBC transfer command
func IBCTransferCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "transfer",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// build the message
			msg, err := buildMsg(from)
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

	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")

	cmd.Flags().String(flagDestChain, "", "Destination chain to send coins")
	cmd.MarkFlagRequired(flagDestChain)
	viper.BindPFlag(flagDestChain, cmd.Flags().Lookup(flagDestChain))
	return cmd
}

func buildMsg(from sdk.AccAddress) (sdk.Msg, error) {
	amount := viper.GetString(flagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return nil, err
	}

    toStr := viper.GetString(flagTo)
	to, err := sdk.AccAddressFromBech32(toStr)
	if err != nil {
		return nil, err
	}
	
	SrcChain := viper.GetString(client.FlagChainID)
	DestChain := viper.GetString(flagDestChain)
	msg := ibc.IBCTransferMsg{
		SrcAddr: from,
		DestAddr: to,
		Coins: coins,
		SrcChain: SrcChain,
		DestChain: DestChain,
	}

	return msg, nil
}
