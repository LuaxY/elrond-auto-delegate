package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/LuaxY/elrond-auto-delegate/config"
	"github.com/LuaxY/elrond-auto-delegate/gas"
	"github.com/LuaxY/elrond-auto-delegate/token"
)

func main() {
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

	fmt.Println("From Address:", fromAddress.Hex())

	if strings.ToLower(cfg.Me.Public) != strings.ToLower(fromAddress.Hex()) {
		log.Fatal("Incorrect address\n", cfg.Me.Public, "\n", fromAddress.Hex())
	}

	client, err := ethclient.Dial(cfg.EthProxy)

	if err != nil {
		log.Fatal(err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Nonce:", nonce)

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

	fmt.Println("Balance:", balance)

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	tx, err := tokenInstance.Approve(auth, common.HexToAddress(cfg.Genesis), balance)

	if err != nil {
		log.Fatal(err)
	}

	j := json.NewEncoder(os.Stdout)
	j.SetIndent("", "  ")
	err = j.Encode(tx)

	if err != nil {
		log.Fatal(err)
	}
}
