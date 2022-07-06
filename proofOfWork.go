package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

// 难度值，表示哈希的前 24 位必须是 0
const targetBits = 24

// 最大块
const maxNonce = math.MaxInt64

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork 将 target 初始化为 1 的大整数，然后左移 256 - targetBits 位，返回 ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		block:  b,
		target: target,
	}

	return pow
}

// prepareData 返回添加了 Nonce 的字节切片
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run 寻找有效哈希
func (pow *ProofOfWork) Run() (int, []byte) {
	// hash 的整形表示
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block")

	for nonce < maxNonce {
		// 准备数据
		data := pow.prepareData(nonce)
		// 用 SHA256 对数据进行哈希
		hash = sha256.Sum256(data)
		// 将哈希转换成一个大整数
		hashInt.SetBytes(hash[:])

		// 将这个大整数与目标进行比较
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
		} else {
			nonce++
		}
	}

	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate 验证哈希是否为小于目标的有效哈希
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
