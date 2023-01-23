package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"go.etcd.io/bbolt"
)

const DB_FILE = "blockchain.db"
const BLOCKS_BUCKET = "blocks"
const GENESIS_COINBASE_DATA = "Blockchain Research Group"

// Blockchain 保存一系列区块
type Blockchain struct {
	tip []byte
	DB  *bbolt.DB
}

// dbExists 检查数据库是否存在
func dbExists() bool {
	if _, err := os.Stat(DB_FILE); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain 创建一个有创世块的区块链
func NewBlockchain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bbolt.Open(DB_FILE, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKS_BUCKET))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	blockchain := Blockchain{tip, db}

	return &blockchain
}

// MineBlock 使用提供的交易挖掘一个新块
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte

	// 在一笔交易被放入一个块之前进行验证：
	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	// BboltDB 只读事务
	// 从数据库中获取最后一个块的哈希，然后用它来挖出一个新的块的哈希
	err := bc.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKS_BUCKET))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	// BboltDB 读写事物
	// 向数据库写入最后一个块的哈希
	err = bc.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKS_BUCKET))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// CreateBlockchain 获取一个地址，该地址将获得挖掘创世块的奖励
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bbolt.Open(DB_FILE, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		cbtx := NewCoinbaseTX(address, GENESIS_COINBASE_DATA)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(BLOCKS_BUCKET))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// Iterator 迭代器
func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{blockchain.tip, blockchain.DB}
}

// FindUTXO 找到所有未花费的交易输出
func (blockchain *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := blockchain.Iterator()

	for {
		block := bci.Next()
		// 对块内的交易进行遍历
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// 查找该地址对应的所有输出
		Outputs:
			for outIndex, out := range tx.VOut {
				// UTXO 未花费的输出意味着这些输出未在任何输入中引用
				// 检查该输出是否已经被包含在一个交易的输入中，检查它是否已经被花费了
				// 跳过那些已经被包含在其他输入中的输出，说明这个输出已经被花费，无法再用了
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIndex {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			// 检查完输出以后，将给定地址所有能够解锁输出的输入聚集起来
			// 能够被地址解锁说明该输入已经被引用了
			// 这不适用于 coinbase 交易，因为它们不解锁输出
			if !tx.IsCoinbase() {
				for _, in := range tx.VIn {
					inTxID := hex.EncodeToString(in.TxID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.VOut)
				}
			}

		}
		// 遍历到创始块则终止
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// FindTransaction 根据交易 ID 查找并返回交易
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// SignTransaction 传入一笔交易，找到它引用的交易，然后对它进行签名
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.VIn {
		prevTX, err := bc.FindTransaction(vin.TxID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction 传入一笔交易，找到它引用的交易，然后对它进行验证
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.VIn {
		prevTX, err := bc.FindTransaction(vin.TxID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
