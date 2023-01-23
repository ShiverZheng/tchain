package blockchain

import (
	"log"

	"go.etcd.io/bbolt"
)

// BlockchainIterator 用于迭代区块链块
type BlockchainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

// Next 从 tip 开始返回链中的下一个块
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKS_BUCKET))
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
