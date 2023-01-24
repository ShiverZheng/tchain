package cli

import (
	"fmt"
	"log"
	"tchain/blockchain"
	"tchain/common"
	"tchain/wallet"
)

func (cli *CLI) getBalance(address string, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := common.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
