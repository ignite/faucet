package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/ignite/cli/v28/ignite/pkg/chaincmd"
	chaincmdrunner "github.com/ignite/cli/v28/ignite/pkg/chaincmd/runner"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosfaucet"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosver"
	"github.com/ignite/cli/v28/ignite/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/faucet/internal/environ"
)

const (
	flagPort            = "port"
	flagKeyringBackend  = "keyring-backend"
	flagSdkVersion      = "sdk-version"
	flagAccountName     = "account-name"
	flagMnemonic        = "mnemonic"
	flagKeyringPassword = "keyring-password"
	flagCLIName         = "cli-name"
	flagDenoms          = "denoms"
	flagCreditAmount    = "credit-amount"
	flagMaxCredit       = "max-credit"
	flagFeeAmount       = "fee-amount"
	flagNode            = "node"
	flagCoinType        = "coin-type"
	flagHome            = "home"
)

const (
	denomSeparator = ","
)

// faucetFlags add the faucet flags to the cobra command.
func faucetFlags(cmd *cobra.Command) {
	cmd.Flags().Int(
		flagPort,
		environ.GetInt("PORT", 8000),
		"tcp port where faucet will be listening for requests",
	)
	cmd.Flags().String(
		flagKeyringBackend,
		environ.GetString("KEYRING_BACKEND", ""),
		"keyring backend to be used",
	)
	cmd.Flags().String(
		flagSdkVersion,
		environ.GetString("SDK_VERSION", cosmosver.Latest.String()),
		"version of sdk",
	)
	cmd.Flags().String(
		flagAccountName,
		environ.GetString("ACCOUNT_NAME", cosmosfaucet.DefaultAccountName),
		"name of the account to be used by the faucet",
	)
	cmd.Flags().String(
		flagMnemonic,
		environ.GetString("MNEMONIC", ""),
		"mnemonic for restoring an account",
	)
	cmd.Flags().String(
		flagKeyringPassword,
		environ.GetString("KEYRING_PASSWORD", ""),
		"password for accessing keyring",
	)
	cmd.Flags().String(
		flagCLIName,
		environ.GetString("CLI_NAME", "gaiad"),
		"name of the cli executable",
	)
	cmd.Flags().String(
		flagDenoms,
		environ.GetString("DENOMS", cosmosfaucet.DefaultDenom),
		"denomination comma separated of the coins sent by default (the first one will be used for the fee denom)",
	)
	cmd.Flags().Uint64(
		flagCreditAmount,
		environ.GetUint64("CREDIT_AMOUNT", cosmosfaucet.DefaultAmount),
		"amount to credit in each request",
	)
	cmd.Flags().Uint64(
		flagMaxCredit,
		environ.GetUint64("MAX_CREDIT", cosmosfaucet.DefaultMaxAmount),
		"maximum credit per account",
	)
	cmd.Flags().Uint64(
		flagFeeAmount,
		environ.GetUint64("FEE_AMOUNT", 0),
		"fee to pay along with the transaction",
	)
	cmd.Flags().String(
		flagNode,
		environ.GetString("NODE", ""),
		"address of tendermint RPC endpoint for this chain",
	)
	cmd.Flags().String(
		flagCoinType,
		environ.GetString("COIN_TYPE", "118"),
		"registered coin type number for HD derivation (BIP-0044), defaults from (satoshilabs/SLIP-0044)",
	)
	cmd.Flags().String(
		flagHome,
		environ.GetString("HOME", ""),
		"replaces the default home used by the chain",
	)
}

func runCmdHandler(cmd *cobra.Command, args []string) error {
	var (
		port, _            = cmd.Flags().GetInt(flagPort)
		keyringBackend, _  = cmd.Flags().GetString(flagKeyringBackend)
		sdkVersion, _      = cmd.Flags().GetString(flagSdkVersion)
		accountName, _     = cmd.Flags().GetString(flagAccountName)
		mnemonic, _        = cmd.Flags().GetString(flagMnemonic)
		keyringPassword, _ = cmd.Flags().GetString(flagKeyringPassword)
		cliName, _         = cmd.Flags().GetString(flagCLIName)
		denoms, _          = cmd.Flags().GetString(flagDenoms)
		creditAmount, _    = cmd.Flags().GetInt64(flagCreditAmount)
		maxCredit, _       = cmd.Flags().GetInt64(flagMaxCredit)
		feeAmount, _       = cmd.Flags().GetInt64(flagFeeAmount)
		node, _            = cmd.Flags().GetString(flagNode)
		coinType, _        = cmd.Flags().GetString(flagCoinType)
		home, _            = cmd.Flags().GetString(flagHome)
	)

	configKeyringBackend, err := chaincmd.KeyringBackendFromString(keyringBackend)
	if err != nil {
		return err
	}

	version, err := cosmosver.Parse(sdkVersion)
	if err != nil {
		return err
	}

	ccoptions := []chaincmd.Option{
		chaincmd.WithKeyringPassword(keyringPassword),
		chaincmd.WithKeyringBackend(configKeyringBackend),
		chaincmd.WithAutoChainIDDetection(),
		chaincmd.WithNodeAddress(node),
		chaincmd.WithVersion(version),
	}

	if home != "" {
		ccoptions = append(ccoptions, chaincmd.WithHome(home))
	}

	cr, err := chaincmdrunner.New(context.Background(), chaincmd.New(cliName, ccoptions...))
	if err != nil {
		return err
	}

	coins := strings.Split(denoms, denomSeparator)
	if len(coins) == 0 {
		return errors.New("empty denoms")
	}

	faucetOptions := []cosmosfaucet.Option{
		cosmosfaucet.Version(version),
		cosmosfaucet.Account(accountName, mnemonic, coinType),
		cosmosfaucet.FeeAmount(sdkmath.NewInt(feeAmount), coins[0]),
	}
	for _, coin := range coins {
		creditAmount := sdkmath.NewInt(creditAmount)
		maxCredit := sdkmath.NewInt(maxCredit)
		faucetOptions = append(faucetOptions, cosmosfaucet.Coin(creditAmount, maxCredit, coin))
	}

	// it is fair to consider the first coin added because it is considered
	// as the default coin during transfer requests.
	faucet, err := cosmosfaucet.New(context.Background(), cr, faucetOptions...)
	if err != nil {
		return err
	}

	http.HandleFunc("/", faucet.ServeHTTP)
	cmd.Printf("listening on :%d", port)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
