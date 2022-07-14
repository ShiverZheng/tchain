package cli

import (
	"fmt"
	"log"
	"tchain/blockchain"
	"tchain/wallet"
)

func (cli *CLI) send(from, to string, amount int) {
	if !wallet.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain(from)
	defer bc.DB.Close()

	tx := blockchain.NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success!")
}
