package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// 奖励金额
const subsidy = 10

// TXInput 交易输入，通过 TxID 和 VOut 两个字段，就可以在区块链上定位到唯一的 UTXO
type TXInput struct {
	TxID      []byte // TxID 引用的 UTXO 所在交易的txID
	VOut      int    // VOut 引用的 UTXO 索引（从 0 开始）
	ScriptSig string // ScriptSig 提供解锁输出 Txid、VOut 的数据
}

// TXOutput 交易输出
type TXOutput struct {
	Value        int    // Value 输出的数量
	ScriptPubKey string // ScriptPubKey 到账地址的锁定脚本
}

// Transaction 交易ID、输入、输出构成一笔交易
type Transaction struct {
	ID   []byte
	VIn  []TXInput
	Vout []TXOutput
}

// SetID 设置交易ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)

	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// NewUTXOTransaction 创建交易
func NewUTXOTransaction(from string, to string, amount int, blockchain *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// 找到足够的未花费输出
	accumulation, validOutputs := blockchain.FindSpendableOutputs(from, amount)

	if accumulation < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for txID, outs := range validOutputs {
		id, err := hex.DecodeString(txID)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{
				TxID:      id,
				VOut:      out,
				ScriptSig: from,
			}
			inputs = append(inputs, input)
		}
	}

	outputs = append(
		outputs,
		TXOutput{
			Value:        amount,
			ScriptPubKey: to,
		})

	// 如果 UTXO 总数超过所需，则产生找零
	if accumulation > amount {
		outputs = append(
			outputs,
			TXOutput{
				Value:        accumulation - amount,
				ScriptPubKey: from,
			},
		)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

// NewCoinbaseTX 创建一个 coinbase 交易
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// IsCoinbase 检查交易是否为发币
func (tx Transaction) IsCoinbase() bool {
	return len(tx.VIn) == 1 && len(tx.VIn[0].TxID) == 0 && tx.VIn[0].VOut == -1
}

// CanUnlockOutputWith 输入锁定方法，检查 Input 的 Vout 是否可以使用提供的地址解锁
func (in *TXInput) CanUnlockOutputWith(unlockingAddress string) bool {
	return in.ScriptSig == unlockingAddress
}

// CanBeUnlockedWith 输出解锁方法，检查地址是否发起了交易
func (out *TXOutput) CanBeUnlockedWith(unlockingAddress string) bool {
	return out.ScriptPubKey == unlockingAddress
}
