package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"go.etcd.io/bbolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Blockchain Research Group"

// Blockchain 保存一系列区块
type Blockchain struct {
	tip []byte
	db  *bbolt.DB
}

// BlockchainIterator 用于迭代区块链块
type BlockchainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

// dbExists 检查数据库是否存在
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain 创建一个有创世块的区块链
func NewBlockchain(address string) *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bbolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
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
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	// BboltDB 只读事务
	// 从数据库中获取最后一个块的哈希，然后用它来挖出一个新的块的哈希
	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	// BboltDB 读写事物
	// 向数据库写入最后一个块的哈希
	err = bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
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
}

// CreateBlockchain 获取一个地址，该地址将获得挖掘创世块的奖励
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bbolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
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
	return &BlockchainIterator{blockchain.tip, blockchain.db}
}

// Next 从 tip 开始返回链中的下一个块
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

// FindUnspentTransactions 返回包含未使用输出的交易列表
func (blockchain *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := blockchain.Iterator()

	for {
		block := bci.Next()
		// 对块内的交易进行遍历
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// 查找该地址对应的所有输出
		Outputs:
			for outIndex, out := range tx.Vout {
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

				// 如果一个输出能够被当前搜索 UTXO 的同一地址锁定
				// 那么这就是需要的输出
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			// 检查完输出以后，将给定地址所有能够解锁输出的输入聚集起来
			// 能够被地址解锁说明该输入已经被引用了
			// 这不适用于 coinbase 交易，因为它们不解锁输出
			if !tx.IsCoinbase() {
				for _, in := range tx.VIn {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TxID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.VOut)
					}
				}
			}

		}
		// 遍历到创始块则终止
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXOs 查找并返回所有未使用的交易输出
func (blockchain *Blockchain) FindUTXOs(address string) []TXOutput {
	var UTXOs []TXOutput
	// 从交易中剥离 UTXO
	unspentTransactions := blockchain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs 查找地址中满足 amount 的 UTXO
func (blockchain *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := blockchain.FindUnspentTransactions(address)
	accumulation := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIndex, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulation < amount {
				accumulation += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIndex)

				if accumulation >= amount {
					break Work
				}
			}
		}
	}

	return accumulation, unspentOutputs
}
