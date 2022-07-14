package cli

import (
	"fmt"
	"log"
	"tchain/blockchain"
	"tchain/wallet"
)

func (cli *CLI) createBlockchain(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address)
	bc.DB.Close()
	fmt.Println("Done!")
}
