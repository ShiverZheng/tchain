package main

type BlockChain struct {
	blocks []*Block
}

// NewBlockChain 创建一个有创世块的链
func NewBlockChain() *BlockChain {
	return &BlockChain{
		blocks: []*Block{NewGenesisBlock()},
	}
}

// AddBlock 向链中加入一个新块
// data 交易数据
func (blockchain *BlockChain) AddBlock(data string) {
	prevBlock := blockchain.blocks[len(blockchain.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	blockchain.blocks = append(blockchain.blocks, newBlock)
}
