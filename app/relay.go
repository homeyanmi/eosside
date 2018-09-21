package cli

import (
	"os"
	"time"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/blockchain-develop/relay-eos-side/types"
	eos "github.com/eoscanada/eos-go"
	"github.com/blockchain-develop/relay-eos-side/x/ibc"
)

const (
	FlagEosChainNode   = "eos-url"
)

type eostransfer struct {
	id int64
	from string
	to string
	memo string
}

type relayCommander struct {
	cdc       *wire.Codec
	//address   sdk.AccAddress
	decoder   auth.AccountDecoder
	mainStore string
	ibcStore  string
	accStore  string

	logger log.Logger
}

// IBC relay command
func IBCRelayCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := relayCommander{
		cdc:       cdc,
		decoder:   authcmd.GetAccountDecoder(cdc),
		ibcStore:  "ibc",
		mainStore: "main",
		accStore:  "acc",

		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}

	cmd := &cobra.Command{
		Use: "relay",
		Run: cmdr.runIBCRelay,
	}

	cmd.Flags().String(FlagEosChainNode, "http://localhost:8888", "<host>:<port>")
	cmd.MarkFlagRequired(FlagEosChainNode)
	viper.BindPFlag(FlagEosChainNode, cmd.Flags().Lookup(FlagEosChainNode))

	return cmd
}

// nolint: unparam
func (c relayCommander) runIBCRelay(cmd *cobra.Command, args []string) {
	eosChainNode := viper.GetString(FlagEosChainNode)
	toChainID := viper.GetString(client.FlagChainID)
	toChainNode := viper.GetString(client.FlagNode)

	c.loop(eosChainNode, toChainID, toChainNode)
}

// This is nolinted as someone is in the process of refactoring this to remove the goto
// nolint: gocyclo
func (c relayCommander) loop(eosChainNode, toChainID, toChainNode string) {

	ctx := context.NewCoreContextFromViper()
	// get password
	passphrase, err := ctx.GetPassphraseFromStdin(ctx.FromAddressName)
	if err != nil {
		panic(err)
	}
	
	//
	eos_api := eos.New("http://127.0.0.1:8888")
	eos_info, _ := eos_api.GetInfo()
	c.logger.Info("eos info", "ServerVersion", eos_info.ServerVersion)
	c.logger.Info("eos info", "HeadBlockNum", eos_info.HeadBlockNum)

	ingressSequenceKey := ibc.IngressSequenceKey()

OUTER:
	for {
		time.Sleep(10 * time.Second)

		//
		ingressSequencebz, err := query(toChainNode, ingressSequenceKey, c.ibcStore)
		if err != nil {
			panic(err)
		}

		var ingressSequence int64
		if ingressSequencebz == nil {
			ingressSequence = 0
		} else if err = c.cdc.UnmarshalBinary(ingressSequencebz, &ingressSequence); err != nil {
			panic(err)
		}
		
		log := fmt.Sprintf("query chain : %s, ingress : %s, number : %d", toChainID, ingressSequenceKey, ingressSequence)
		c.logger.Info("log", "string", log)

		//
		gettable_request := eos.GetTableRowsRequest {
			Code: "pegzone",
			Scope: "pegzone",
			Table: "actioninfo",
			LowerBound: "1",
		}

		gettable_response, eos_err := eos_api.GetTableRows(gettable_request)
		if eos_err != nil {
			c.logger.Info("eos get table failed", "error", eos_err)
			panic("eos get table failed")
		}

		c.logger.Info("eos get table successful", "result", gettable_response.Rows)
		
		//
		var transfers []*eostransfer
		err_json := gettable_response.BinaryToStructs(&transfers)
		if err_json != nil {
			c.logger.Info("eos get table failed", "error", err_json)
			panic("eos get table failed")
		}


		/*
		err_json := c.cdc.UnmarshalJSON(gettable_response.Rows, &transfers)
		if err_json != nil {
			panic("eos get table failed")
		}
		*/
		
		c.logger.Info("query chain table", "total", len(transfers))

		seq := (c.getSequence(toChainNode))
		//c.logger.Info("broadcast tx seq", "number", seq)

		for i, _ := range transfers {
			
			//
			ibc_msg := ibc.Transfer{
				//SrcAddr: fromChainID,
				//DestAddr: toChainID,
				//Coins: from,
			}
			
			bz, err_json := c.cdc.MarshalBinary(ibc_msg)
			if err_json != nil {
				panic(err)
			}
			
			
			// get the from address
			from, err := ctx.GetFromAddress()
			if err != nil {
				panic(err) 
			}
			
			relay_msg := ibc.IBCRelayMsg {
            	PayloadType: ibc.TRANSFER,
            	Payload: bz,
            	Sequence: int64(i),
            	Relayer: from,
            }
			
			new_ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(c.cdc)).WithSequence(int64(i))
			//ctx = ctx.WithNodeURI(viper.GetString(FlagHubChainNode))
			//ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(c.cdc))
			new_ctx, err = new_ctx.Ensure(ctx.FromAddressName)
			new_ctx = new_ctx.WithSequence(seq)
			xx_res, err := new_ctx.SignAndBuild(ctx.FromAddressName, passphrase, []sdk.Msg{relay_msg}, c.cdc)
			if err != nil {
				panic(err)
			}

			//
			c.logger.Info("broadcast tx, type : newibcpacket, sequence : ", "int", seq)
			
			//
			err = c.broadcastTx(seq, toChainNode, xx_res)
			seq++
			if err != nil {
				c.logger.Error("error broadcasting ingress packet", "err", err)
				continue OUTER // TODO replace to break, will break first loop then send back to the beginning (aka OUTER)
			}

			c.logger.Info("Relayed IBC packet", "index", i)
		}
	}
}

func (c relayCommander) getaddress() (addr sdk.AccAddress) {
	address, err := context.NewCoreContextFromViper().GetFromAddress()
	if err != nil {
		panic(err)
	}
	return address
}

func query(node string, key []byte, storeName string) (res []byte, err error) {
	return context.NewCoreContextFromViper().WithNodeURI(node).QueryStore(key, storeName)
}

func queryWithProof(node string, key []byte, storeName string) (x abci.ResponseQuery, err error) {
	return context.NewCoreContextFromViper().WithNodeURI(node).QueryStoreWithProof(key, storeName)
}

func commitWithProof(node string, height int64, storeName string) (x *ctypes.ResultCommit, err error) {
	return context.NewCoreContextFromViper().WithNodeURI(node).CommitWithProof(storeName, height)
}

func (c relayCommander) broadcastTx(seq int64, node string, tx []byte) error {
	_, err := context.NewCoreContextFromViper().WithNodeURI(node).WithSequence(seq + 1).BroadcastTx(tx)
	return err
}

func (c relayCommander) getSequence(node string) int64 {
	res, err := query(node, auth.AddressStoreKey(c.getaddress()), c.accStore)
	if err != nil {
		panic(err)
	}
	if nil != res {
		account, err := c.decoder(res)
		if err != nil {
			panic(err)
		}

		return account.GetSequence()
	}

	return 0
}
