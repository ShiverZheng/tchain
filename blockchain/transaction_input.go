package blockchain

import (
	"bytes"
	"tchain/wallet"
)

// TXInput 交易输入，通过 TxID 和 VOut 两个字段，就可以在区块链上定位到唯一的 UTXO
type TXInput struct {
	TxID      []byte // TxID 引用的 UTXO 所在交易的txID
	VOut      int    // VOut 引用的 UTXO 索引（从 0 开始）
	Signature []byte // Signature 签名 替代 ScriptSig 脚本提供解锁输出 Txid、VOut 的数据
	PubKey    []byte // PubKey 公钥
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)

	return bytes.Equal(lockingHash, pubKeyHash)
}

// CanUnlockOutputWith 输入锁定方法，检查 Input 的 Vout 是否可以使用提供的地址解锁
func (in *TXInput) CanUnlockOutputWith(unlockingAddress string) bool {
	// return in.ScriptSig == unlockingAddress
	return true
}
