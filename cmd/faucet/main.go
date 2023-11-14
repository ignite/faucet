package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	sdkmath "cosmossdk.io/math"
	log "github.com/sirupsen/logrus"

	"github.com/ignite/cli/ignite/pkg/chaincmd"
	chaincmdrunner "github.com/ignite/cli/ignite/pkg/chaincmd/runner"
	"github.com/ignite/cli/ignite/pkg/cosmosfaucet"
	"github.com/ignite/cli/ignite/pkg/cosmosver"
)

func main() {
	flag.Parse()

	configKeyringBackend, err := chaincmd.KeyringBackendFromString(keyringBackend)
	if err != nil {
		log.Fatal(err)
	}

	ccoptions := []chaincmd.Option{
		chaincmd.WithKeyringPassword(keyringPassword),
		chaincmd.WithKeyringBackend(configKeyringBackend),
		chaincmd.WithAutoChainIDDetection(),
		chaincmd.WithNodeAddress(nodeAddress),
	}

	if home != "" {
		ccoptions = append(ccoptions, chaincmd.WithHome(home))
	}

	ccoptions = append(ccoptions,
		chaincmd.WithVersion(cosmosver.Latest),
	)

	cr, err := chaincmdrunner.New(context.Background(), chaincmd.New(appCli, ccoptions...))
	if err != nil {
		log.Fatal(err)
	}

	coins := strings.Split(defaultDenoms, denomSeparator)

	faucetOptions := make([]cosmosfaucet.Option, len(coins))
	for i, coin := range coins {
		creditAmount := sdkmath.NewInt(int64(creditAmount))
		maxCredit := sdkmath.NewInt(int64(maxCredit))

		faucetOptions[i] = cosmosfaucet.Coin(creditAmount, maxCredit, coin)
	}

	faucetOptions = append(faucetOptions, cosmosfaucet.Account(keyName, keyMnemonic, coinType))
	// it is fair to consider the first coin added because it is considered as the default coin during transfer requests.
	faucetOptions = append(faucetOptions, cosmosfaucet.FeeAmount(sdkmath.NewInt(int64(feeAmount)), coins[0]))

	faucet, err := cosmosfaucet.New(context.Background(), cr, faucetOptions...)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", faucet.ServeHTTP)
	log.Infof("listening on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
