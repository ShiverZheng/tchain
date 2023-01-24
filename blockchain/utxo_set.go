package blockchain

import (
	"encoding/hex"
	"log"

	"go.etcd.io/bbolt"
)

const UTXO_BUCKET = "chainstate"

// UTXOSet UTXO 集合
type UTXOSet struct {
	Blockchain *Blockchain
}

// FindSpendableOutputs 查找并返回 UTXO 在输入中的引用
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.DB

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_BUCKET))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO 为公钥哈希找到 UTXO
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.DB

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_BUCKET))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// Reindex 初始化 UTXO 集
func (u UTXOSet) Reindex() {
	db := u.Blockchain.DB
	bucketName := []byte(UTXO_BUCKET)

	// 如果 bucket 存在就先移除
	err := db.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bbolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	// 然后从区块链中获取所有的未花费输出
	UTXO := u.Blockchain.FindUTXO()

	// 最终将输出保存到 bucket 中
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

}

// Update 当挖出一个新块时，更新 UTXO 集，使其保持 UTXO 集处于最新状态，并且存储最新交易的输出
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.DB

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_BUCKET))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.VIn {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.TxID)
					outs := DeserializeOutputs(outsBytes)

					// 从新挖出来的交易中加入 UTXO
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.VOut {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					// 如果一笔交易的输出被移除，并且不再包含任何输出，那么这笔交易也应该被移除。
					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.TxID)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.TxID, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.VOut {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// CountTransactions 返回 UTXO 集中的交易数量
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.DB
	counter := 0

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(UTXO_BUCKET))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}
