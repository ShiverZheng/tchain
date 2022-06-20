package main

import "fmt"

func main() {
	blockchian := NewBlockChain()

	blockchian.AddBlock("Send 100 FX to Blob")
	blockchian.AddBlock("Send 500 FX to Mary")

	for _, block := range blockchian.blocks {
		fmt.Printf("Prev Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
