package main

import (
	"flag"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosfaucet"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosver"

	"github.com/tendermint/faucet/internal/environ"
)

const (
	denomSeparator = ","
)

var (
	port            int
	keyringBackend  string
	sdkVersion      string
	keyName         string
	keyMnemonic     string
	keyringPassword string
	appCli          string
	defaultDenoms   string
	creditAmount    uint64
	maxCredit       uint64
	feeAmount       uint64
	nodeAddress     string
	coinType        string
	home            string
)

func init() {
	flag.IntVar(&port, "port",
		environ.GetInt("PORT", 8000),
		"tcp port where faucet will be listening for requests",
	)
	flag.StringVar(&keyringBackend, "keyring-backend",
		environ.GetString("KEYRING_BACKEND", ""),
		"keyring backend to be used",
	)
	flag.StringVar(&sdkVersion, "sdk-version",
		environ.GetString("SDK_VERSION", cosmosver.Latest.String()),
		"version of sdk",
	)
	flag.StringVar(&keyName, "account-name",
		environ.GetString("ACCOUNT_NAME", cosmosfaucet.DefaultAccountName),
		"name of the account to be used by the faucet",
	)
	flag.StringVar(&keyMnemonic, "mnemonic",
		environ.GetString("MNEMONIC", ""),
		"mnemonic for restoring an account",
	)
	flag.StringVar(&keyringPassword, "keyring-password",
		environ.GetString("KEYRING_PASSWORD", ""),
		"password for accessing keyring",
	)
	flag.StringVar(&appCli, "cli-name",
		environ.GetString("CLI_NAME", "gaiad"),
		"name of the cli executable",
	)
	flag.StringVar(&defaultDenoms, "denoms",
		environ.GetString("DENOMS", cosmosfaucet.DefaultDenom),
		"denomination comma separated of the coins sent by default (the first one will be used for the fee denom)",
	)
	flag.Uint64Var(&creditAmount,
		"credit-amount",
		environ.GetUint64("CREDIT_AMOUNT", cosmosfaucet.DefaultAmount),
		"amount to credit in each request",
	)
	flag.Uint64Var(&maxCredit,
		"max-credit", environ.GetUint64("MAX_CREDIT", cosmosfaucet.DefaultMaxAmount),
		"maximum credit per account",
	)
	flag.Uint64Var(&feeAmount,
		"fee-amount", environ.GetUint64("FEE_AMOUNT", 0),
		"fee to pay along with the transaction",
	)
	flag.StringVar(&nodeAddress, "node",
		environ.GetString("NODE", ""),
		"address of tendermint RPC endpoint for this chain",
	)
	flag.StringVar(&coinType, "coin-type",
		environ.GetString("COIN_TYPE", "118"),
		"registered coin type number for HD derivation (BIP-0044), defaults from (satoshilabs/SLIP-0044)",
	)
	flag.StringVar(&home, "home",
		environ.GetString("HOME", ""),
		"replaces the default home used by the chain",
	)
}
