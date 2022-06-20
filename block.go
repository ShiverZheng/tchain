package main

import (
	"time"
)

// Block 由区块头和交易两部分构成
// Timestamp, PrevBlockHash, Hash 属于区块头（block header）
// Timestamp     : 当前时间戳，也就是区块创建的时间
// PrevBlockHash : 前一个块的哈希
// Hash          : 当前块的哈希
// Data          : 区块实际存储的交易信息
type Block struct {
	Timestamp     int64
	PrevBlockHash []byte
	Hash          []byte
	Data          []byte
	Nonce         int
}

// NewBlock 用于生成新块
// 基于 Data 和 PrevBlockHash 计算得到当前块的哈希
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Data:          []byte(data),
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock 生成创世块
func NewGenesisBlock() *Block {
	return NewBlock("Gensis Block", []byte{})
}
