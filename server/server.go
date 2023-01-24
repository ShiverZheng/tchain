package server

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"tchain/blockchain"
)

const PROTOCOL = "tcp"
const NODE_VERSION = 1
const COMMAND_LENGTH = 12

var nodeAddress string
var miningAddress string
var KnownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]blockchain.Transaction)

type addr struct {
	AddrList []string
}

type version struct {
	Version    int
	BestHeight int    // 区块链中节点的高
	AddrFrom   string // 发送者的地址
}

type block struct {
	AddrFrom string
	Block    []byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type getBlocks struct {
	AddrFrom string
}

type getData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

// 向其他节点展示当前节点的块和交易
type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

func commandToBytes(command string) []byte {
	var bytes [COMMAND_LENGTH]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(PROTOCOL, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendBlock(addr string, b *blockchain.Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getData{nodeAddress, kind, id})
	request := append(commandToBytes("getData"), payload...)

	sendData(address, request)
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func SendTx(addr string, tnx *blockchain.Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request)
}

func sendGetBlocks(address string) {
	payload := gobEncode(getBlocks{nodeAddress})
	request := append(commandToBytes("getBlocks"), payload...)

	sendData(address, request)
}

func requestBlocks() {
	for _, node := range KnownNodes {
		sendGetBlocks(node)
	}
}

func nodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func handleVersion(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload version

	buff.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	// BestHeight 与自身进行比较
	if myBestHeight < foreignerBestHeight {

		// 消息中的区块链更长发送 getBlocks 消息
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		// 自身节点的区块链更长
		// 回复 version 消息
		sendVersion(payload.AddrFrom, bc)
	}

	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}
}

// 当接收到一个新块时，我们把它放到区块链里面。
func handleBlock(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := blockchain.DeserializeBlock(blockData)

	fmt.Println("Received a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	// 如果还有更多的区块需要下载，继续从上一个下载的块的那个节点继续请求
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]

		// 当最后把所有块都下载完后，对 UTXO 集进行重新索引
	} else {
		UTXOSet := blockchain.UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func handleTx(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	// 将新交易放到内存池
	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			// 检查当前节点是否是中心节点
			// 在这里中心节点并不会挖矿
			// 它只会将新的交易推送给网络中的其他节点
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		// miningAddress 只会在矿工节点上设置
		// 如果当前节点（矿工）的内存池中有两笔或更多的交易，开始挖矿：
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*blockchain.Transaction

			// 内存池中所有交易都是通过验证的
			// 无效的交易会被忽略
			for id := range mempool {
				tx := mempool[id]
				// 验证后的交易被放到一个块里
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			// 如果没有有效交易，则挖矿中断
			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			// 同时还有附带奖励的 coinbase 交易
			cbTx := blockchain.NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := blockchain.UTXOSet{bc}
			// 当块被挖出来以后，UTXO 集会被重新索引
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			// 当一笔交易被挖出来以后就会被从内存池中移除
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			// 当前节点所连接到的所有其他节点，接收带有新块哈希的 inv 消息
			for _, node := range KnownNodes {
				if node != nodeAddress {
					// 在处理完消息后，它们可以对块进行请求。
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleGetBlocks(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload getBlocks

	buff.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleInv(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetData(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload getData

	buff.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		SendTx(payload.AddrFrom, &tx)
	}
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[COMMAND_LENGTH:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	KnownNodes = append(KnownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(KnownNodes))
	requestBlocks()
}

func handleConnection(conn net.Conn, bc *blockchain.Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:COMMAND_LENGTH])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getBlocks":
		handleGetBlocks(request, bc)
	case "getData":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

// StartServer starts a node
func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(PROTOCOL, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := blockchain.NewBlockchain(nodeID)

	if nodeAddress != KnownNodes[0] {
		sendVersion(KnownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

func sendVersion(addr string, bc *blockchain.Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(version{NODE_VERSION, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}
