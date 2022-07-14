package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"tchain/wallet"
)

// 奖励金额
const subsidy = 10

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

// Serialize 返回序列化的交易
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash 返回交易的哈希
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// NewUTXOTransaction 创建交易
func NewUTXOTransaction(from string, to string, amount int, blockchain *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	wallets, err := wallet.NewWallets()
	if err != nil {
		log.Panic(err)
	}

	wlt := wallets.GetWallet(from)
	pubKeyHash := wallet.HashPubKey(wlt.PublicKey)

	// 找到足够的未花费输出
	accumulation, validOutputs := blockchain.FindSpendableOutputs(pubKeyHash, amount)

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
				Signature: nil,
				PubKey:    wlt.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	outputs = append(
		outputs,
		*NewTXOutput(amount, to),
	)

	// 如果 UTXO 总数超过所需，则产生找零
	if accumulation > amount {
		outputs = append(outputs, *NewTXOutput(accumulation-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	blockchain.SignTransaction(&tx, wlt.PrivateKey)

	return &tx
}

// NewCoinbaseTX 创建一个 coinbase 交易
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// IsCoinbase 检查交易是否为发币
func (tx Transaction) IsCoinbase() bool {
	return len(tx.VIn) == 1 && len(tx.VIn[0].TxID) == 0 && tx.VIn[0].VOut == -1
}

// TrimmedCopy 返回包括原始交易中所有的 TXI 和 TXO 的副本，其中 TXI 的 Signature, PubKey 设置为 nil
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.VIn {
		inputs = append(inputs, TXInput{vin.TxID, vin.VOut, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Sign 对每一个交易输入进行签名
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	// Coinbase 交易没有真实的 TXI，因此这笔交易不进行签名
	if tx.IsCoinbase() {
		return
	}

	// 对交易进行签名时，需要获取该交易所有 TXI 引用的 TXO 列表，因此需要存储这些 TXO 对应的交易
	for _, vin := range tx.VIn {
		if prevTXs[hex.EncodeToString(vin.TxID)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	// 接下来对TrimmedCopy交易中所有的TXI进行遍历：
	for inID, vin := range txCopy.VIn {
		prevTx := prevTXs[hex.EncodeToString(vin.TxID)]
		// 对于每个 TXI 为了确保万无一失，再次将 Signature 字段设置为 nil
		txCopy.VIn[inID].Signature = nil
		// 将 TXI 的 PubKey 设置为其所引用的 TXO 的 PubKeyHash
		txCopy.VIn[inID].PubKey = prevTx.Vout[vin.VOut].PubKeyHash
		// Hash 方法将交易序列化并通过 SHA-256 进行哈希，生成的哈希结果就是待签名的数据
		txCopy.ID = txCopy.Hash()
		// 将 PubKey 字段重新设置为 nil 避免影响后续的迭代
		txCopy.VIn[inID].PubKey = nil

		// 使用 ECDSA 签名算法通过私钥 privKey 对 txCopy.ID 进行签名，生成一对数字序列
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		// 一个 ECDSA 签名就是一对数字，将这对数字连接起来，并存储在输入的 Signature 字段
		signature := append(r.Bytes(), s.Bytes()...)

		tx.VIn[inID].Signature = signature
	}
}

// Verify 对交易的签名进行验证
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()

	// 创建椭圆曲线用于生成键值对
	curve := elliptic.P256()

	// 对于每个 TXI 的签名进行验证
	for inID, vin := range tx.VIn {
		// 这个过程和 Sign 方法是一致的，因为验证的数据需要和签名的数据是一致的
		prevTx := prevTXs[hex.EncodeToString(vin.TxID)]
		txCopy.VIn[inID].Signature = nil
		txCopy.VIn[inID].PubKey = prevTx.Vout[vin.VOut].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.VIn[inID].PubKey = nil

		// 解包存储在 TXInput.Signature 和 TXInput.PubKey 中的值，因为一个签名就是一对数字，一个公钥就是一对坐标
		// 之前为了存储将它们连接在一起，现在我们需要对它们进行解包在 crypto/ecdsa 函数中使用
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		// 从输入提取的公钥创建一个 ecdsa.PublicKey
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}

		// 通过传入输入中提取的签名执行 ecdsa.Verify
		// 如果所有的输入都被验证，返回 true
		// 如果有任何一个验证失败，返回 false
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}
