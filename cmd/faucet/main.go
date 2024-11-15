package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	sdkmath "cosmossdk.io/math"
	log "github.com/sirupsen/logrus"

	"github.com/ignite/cli/v28/ignite/pkg/chaincmd"
	chaincmdrunner "github.com/ignite/cli/v28/ignite/pkg/chaincmd/runner"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosfaucet"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosver"
)

func main() {
	flag.Parse()

	configKeyringBackend, err := chaincmd.KeyringBackendFromString(keyringBackend)
	if err != nil {
		log.Fatal(err)
	}

	version, err := cosmosver.Parse(sdkVersion)
	if err != nil {
		log.Fatal(err)
	}

	ccoptions := []chaincmd.Option{
		chaincmd.WithKeyringPassword(keyringPassword),
		chaincmd.WithKeyringBackend(configKeyringBackend),
		chaincmd.WithAutoChainIDDetection(),
		chaincmd.WithNodeAddress(nodeAddress),
		chaincmd.WithVersion(version),
	}

	if home != "" {
		ccoptions = append(ccoptions, chaincmd.WithHome(home))
	}

	ccoptions = append(ccoptions,
		chaincmd.WithVersion(version),
	)

	cr, err := chaincmdrunner.New(context.Background(), chaincmd.New(appCli, ccoptions...))
	if err != nil {
		log.Fatal(err)
	}

	coins := strings.Split(defaultDenoms, denomSeparator)
	if len(coins) == 0 {
		log.Fatal("empty denoms")
	}

	faucetOptions := []cosmosfaucet.Option{
		cosmosfaucet.Version(version),
		cosmosfaucet.Account(keyName, keyMnemonic, coinType),
		cosmosfaucet.FeeAmount(sdkmath.NewInt(int64(feeAmount)), coins[0]),
	}
	for _, coin := range coins {
		creditAmount := sdkmath.NewInt(int64(creditAmount))
		maxCredit := sdkmath.NewInt(int64(maxCredit))
		faucetOptions = append(faucetOptions, cosmosfaucet.Coin(creditAmount, maxCredit, coin))
	}

	// it is fair to consider the first coin added because it is considered as the default coin during transfer requests.

	faucet, err := cosmosfaucet.New(context.Background(), cr, faucetOptions...)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", faucet.ServeHTTP)
	log.Infof("listening on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
