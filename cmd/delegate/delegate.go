package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/LuaxY/elrond-auto-delegate/config"
	"github.com/LuaxY/elrond-auto-delegate/gas"
	"github.com/LuaxY/elrond-auto-delegate/genesis"
	"github.com/LuaxY/elrond-auto-delegate/token"
)

func main() {
	ctx := context.Background()

	var configFile string
	flag.StringVar(&configFile, "c", "mainnet.yml", "configuration file")
	flag.Parse()

	cfg, err := config.NewConfig(configFile)

	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA(cfg.Me.Private)

	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	log.Println("From Address:", fromAddress.Hex())

	if strings.ToLower(cfg.Me.Public) != strings.ToLower(fromAddress.Hex()) {
		log.Fatal("Incorrect address\n", cfg.Me.Public, "\n", fromAddress.Hex())
	}

	client, err := ethclient.Dial(cfg.EthProxy)

	if err != nil {
		log.Fatal(err)
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddress)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Nonce:", nonce)

	gasPrice, err := gas.GetPrice(gas.Fastest)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Current Gas Price:", gasPrice)

	tokenInstance, err := token.NewToken(common.HexToAddress(cfg.Token), client)

	if err != nil {
		log.Fatal(err)
	}

	balance, err := tokenInstance.BalanceOf(nil, common.HexToAddress(cfg.Me.Public))

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Balance:", balance)

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = 0
	auth.GasPrice = gasPrice

	genesisInstance, err := genesis.NewGenesis(common.HexToAddress(cfg.Genesis), client)

	if err != nil {
		log.Fatal(err)
	}

	delegationAmountLimit, err := genesisInstance.DelegationAmountLimit(nil)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Limit:  ", delegationAmountLimit)

	sink := make(chan *genesis.GenesisStakeWithdrawn)
	_, err = genesisInstance.WatchStakeWithdrawn(nil, sink, nil)

	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(15 * time.Second)

	for {
		currentTotalDelegated, err := genesisInstance.CurrentTotalDelegated(nil)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("Current:", currentTotalDelegated)

		if currentTotalDelegated.Cmp(delegationAmountLimit) == -1 {
			delta := big.NewInt(0).Sub(delegationAmountLimit, currentTotalDelegated)
			log.Println("Available:", delta)

			delegate := big.NewInt(0)

			if balance.Cmp(delta) == 1 {
				delegate = delta
			} else {
				delegate = balance
			}

			log.Println("Delegate:", delegate)

			hash, err := hex.DecodeString(cfg.Me.Elrond)

			if err != nil {
				log.Fatal(err)
			}

			var elrondHash [32]byte
			copy(elrondHash[:], hash[:32])

			gasPrice, err = gas.GetPrice(gas.Fastest)

			if err == nil {
				auth.GasPrice = gasPrice
				log.Println("New Gas Price:", auth.GasPrice)
			} else {
				log.Println("failed to update gas price, use previous one", auth.GasPrice)
			}

			tx, err := genesisInstance.DepositDelegateStake(auth, delegate, elrondHash)

			if err != nil {
				log.Fatal(err)
			}

			j := json.NewEncoder(os.Stdout)
			j.SetIndent("", "  ")
			err = j.Encode(tx)

			if err != nil {
				log.Fatal(err)
			}

			log.Println("Tokens delegated")
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			continue
		case withdrawn := <-sink:
			log.Println(withdrawn.Account.Hex(), withdrawn.Amount)
		}
	}
}
