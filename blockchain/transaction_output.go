package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"tchain/common"
)

// TXOutput 交易输出
type TXOutput struct {
	Value      int    // Value 输出的数量
	PubKeyHash []byte // PubKeyHash 公钥Hash
}

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}

// Lock 对 TXO 进行加锁，从地址中从解码出哈希后的公钥，将其保存到PubKeyHash中
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := common.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey 输出解锁方法，检查地址是否发起了交易
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

// NewTXOutput 创建一盒新的 TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
