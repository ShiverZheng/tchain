package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// Block 由区块头和交易两部分构成
// Timestamp, PrevBlockHash, Hash 属于区块头
type Block struct {
	Timestamp     int64  // 当前时间戳，也就是区块创建的时间
	PrevBlockHash []byte // 前一个块的哈希
	Hash          []byte // 当前块的哈希
	Data          []byte // 区块实际存储的交易信息
	Nonce         int
}

// Serialize 序列化区块
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)

	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock 反序列化区块
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewBuffer(d))
	err := decoder.Decode(&block)

	if err != nil {
		log.Panic(err)
	}

	return &block
}

// NewBlock 基于 Data 和 PrevBlockHash 计算得到当前块的哈希，创建并返回区块
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
