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
	defer bc.DB.Close()

	// 当一个新的区块链被创建以后，就会立刻进行重建索引
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
