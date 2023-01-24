package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"tchain/merkle"
	"time"
)

// Block 由区块头和交易两部分构成
// Timestamp, PrevBlockHash, Hash 属于区块头
type Block struct {
	Timestamp     int64          // 当前时间戳，也就是区块创建的时间
	PrevBlockHash []byte         // 前一个块的哈希
	Hash          []byte         // 当前块的哈希
	Transactions  []*Transaction // 区块实际存储的交易信息
	Nonce         int            // 随机数
	Height        int            // 块的高度k
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

// HashTransactions 返回块中包含的交易的哈希
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		// 交易被序列化（使用 encoding/gob）
		transactions = append(transactions, tx.Serialize())
	}
	// 使用序列后的交易构建一个 Merkle 树
	mTree := merkle.NewMerkleTree(transactions)

	// 树根将会作为块交易的唯一标识符
	return mTree.RootNode.Data
}

// NewBlock 基于 Data 和 PrevBlockHash 计算得到当前块的哈希，创建并返回区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Transactions:  transactions,
		Nonce:         0,
		Height:        height,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock 生成创世块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}
