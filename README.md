# TChain


## 工作原理

### 交易

交易是加密货币系统中最重要的部分，加密货币中的其他一切都是为了确保交易可以被创建、在网络上传播、验证，并最终添加到全局交易分类账本（区块链）中。

> 以下内容以比特币的实现为参考

相较于基于账户余额的记账方式比特币的实现则完全不同：

- 没有账户
- 没有余额
- 没有地址信息
- 没有货币信息
- 没有付款人、收款人

由于区块链是完全开放的，且匿名的，账户中不包含任何交易额信息，交易也不会将钱从一个账户转到另一个账户，也不存储任何账户余额信息。

仅有的就是交易信息，交易信息到底有什么呢？

这是一个交易的数据，接下来将逐步剖析其中的含义。
```json
{
  "version": 1,
  "locktime": 0,
  "vin": [
    {
      "txid":"7957a35fe64f80d234d76d83a2a8f1a0d8149a41d81de548f0a65a8a999f6f18",
      "vout": 0,
      "scriptSig": "3045022100884d142d86652a3f47ba4746ec719bbfbd040a570b1deccbb6498c75c4ae24cb02204b9f039ff08df09cbe9f6addac960298cad530a863ea8f53982c09db8f6e3813[ALL] 0484ecc0d46f1918b30928fa0e4ed99f16a0fb4fde0735e7ade8416ab9fe423cc5412336376789d172787ec3457eee41c04f4938de5cc17b4a10fa336a8d752adf",
      "sequence": 4294967295
    }
 ],
  "vout": [
    {
      "value": 0.01500000,
      "scriptPubKey": "OP_DUP OP_HASH160 ab68025513c3dbd2f7b92a94e0581f5d50f654e7 OP_EQUALVERIFY OP_CHECKSIG"
    },
    {
      "value": 0.08450000,
      "scriptPubKey": "OP_DUP OP_HASH160 7f9b1a7fb68d60c536c2fd8aeaa53a8f3cc025a8 OP_EQUALVERIFY OP_CHECKSIG",
    }
  ]
}
```

#### UTXO
> You never actually own a Patek Philippe , you merely look after it for the next generation.
> 
> ——没有人能真正拥有百达翡丽，只不过为下一代保管而已。

将百达翡丽流传最广的传达了产品价值的广告语运用到比特币里，

> 没有人能真正拥有比特币，只不过为别人保管而已。

但这里并不是想表达比特币的价值，而是想引出比特币的交易模型：

UTXO (Unspent Transaction Output)，未花费的输出。

初看这个名词有些不知所云，可以姑且将它当成现实世界中的现金，假设一个商品的价格是90元，小张想要购买，现在他手头有100元、50元、20元、20元四张纸币。

现实世界中是没有90元面值的纸币，小张也不可能将一张100元纸币撕个九成出来进行支付。这个时候小张可以使用一张50元、两张20元进行支付，也可以使用一张100元进行支付，同时拿回商家找零的10元。

如果把交易场景切换到UTXO模型中，无论是50元、20元、100元还是找零的10元，都可以将其视为UTXO。由于小张没有90元的UTXO，因此小张可以使用多笔面值小的UTXO进行付款，可以视作当下这个交易的交易输入。

当然也可以输入一笔大的UTXO，商家接过钱，这时候对于小张来说这些钱就变成了交易输出，变成了商家的交易输入，还应该有找零，但注意这时候不再是商家找零小张10元，UTXO模型自动完成了这一步骤，它将这一笔大的UTXO拆分开，一部分输出给商家，一部分输出给小张。现在对于商家来说他有了新的90元的UTXO，小张有了新的10元的UTXO。

为什么说是新的UTXO，再回看一下UTXO名称的含义——未花费的交易输出，那么一旦消费过了，就不再是UTXO。

若小张使用一张50元和两张20元进行支付，那这三笔UTXO会变成“已花费”，商家得到一个90元的新UTXO，如果小张使用100元进行支付，那这笔100元的UTXO就变成了“已花费”，商家得到90元的新UTXO、小张得到10元新UTXO。

当商家想要向厂家进货时，他需要向厂家支付1000元的货款，为了实现这笔交易，商家需要消费由小张购物产生的交易所产生的90元的UTXO，再加上其他顾客购物产生的UTXO，而厂家不想要收到黑钱，厂家就需要去查商家的账，看看这些UTXO是不是商家卖货得来的，厂家也不用把所有的帐都查一遍，只要查够1000元的账，足够完成这笔进货交易就行了，厂家查到这些UTXO都是由商家卖货交易产生的时候，这些UTXO变成了进货这笔交易的输入，那么厂家就得到了由这笔交易创建的1000元的新的UTXO。

商家也不想收到黑钱，所以商家也会去查顾客的账，过程是一样的，就不再这里赘述，其实查账这个过程不需要任何收款方来进行，UTXO模型会去做这个事情。在UTXO的模型里，想要交易就需要向前追溯UTXO的来源确定这个UTXO是否存在，一笔交易会消耗先前的存在的UTXO，并创建新的UTXO以备未来的交易消耗。通过这种方式，一定数量的比特币价值在不同所有者之间转移，并在交易链中消耗和创建UTXO。

*被交易消费的UTXO被称为交易输入，由交易创建的UTXO被称为交易输出。每一笔交易的来源(输入)都来自于上一笔交易的输出。UTXO交易过程跟踪的是每个Token的所有权的转移。*


所以没有人能真正拥有比特币，你看到的地址账户里的比特币，是”钱包“帮你在UTXO模型里算出来的，其实它的“价值”还躺在上一笔交易里呢！

#### Coinbase
对于输出与输入链来说，有一种特殊的交易，称为“币基交易”，它是每个区块的第一笔交易。
这种交易是由获胜的矿工放置的，给这个矿工创建了新的可花费比特币，它是挖矿的奖励。
这个特殊的币基交易不消费UTXO，它有一个特殊类型的输入，称为“币基”。
这就是挖矿过程中创建比特币货币供应的方式。
> 输入和输出，哪个先产生？先有鸡还是先有蛋？
> 严格来讲，先产生输出，因为币基交易（创造新比特币）没有输入，它是无中生有。


### 地址

[1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa](https://explorer.btc.com/btc/address/1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa)

这是世界上首个比特币地址，据说属于比特币发明人中本聪。

比特币地址是公开的，如果你想转给某人一些BTC，那么就需要知道其地址。

这些地址并不代表钱包，仅仅是具备可读格式的公钥。

在比特币世界中，你的ID是一个密钥对（公/私钥），该密钥对需要保存在你的电脑或者其他你可以直接存取到它的设备。

比特币通过密码学算法来创建密钥对，从而保证在不能直接存取该密钥情况下，没人可以动你的钱。

公钥加密算法使用一对密钥：公钥和私钥。

公钥可以给其他人，而私钥不应该给其他人，你的私钥就代表你本人。

本质上，比特币钱包就是一个密钥对。当你安装钱包客户端或者比特币客户端创建一个新地址时，一个密钥对将会被创建。在比特币世界，谁拥有密钥，谁就可以掌控属于该密钥的钱。

公钥和私钥仅仅是一些随机字节序列，人们无法直接读取，因此比特币使用一个算法用于将公钥转换成可读的字符串——[助记词](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki)。

## Development

### 区块

```go
type Block struct {
	Timestamp     int64     // 当前时间戳，也就是区块创建的时间
	PrevBlockHash []byte    // 前一个块的哈希
	Hash          []byte    // 当前块的哈希
	Data          []byte    // 区块实际存储的交易信息
	Nonce         int
}
```

...

### 持久化
使用 [bbolt](https://github.com/etcd-io/bbolt) 对区块链进行持久化

#### 数据库结构
> Bitcoin Core 使用两个 `bucket` 来存储数据：
> - `blocks` 存储描述一条链中所有块的元数据
> - `chainstate` 存储一条链的状态，也就是当前所有的未花费的交易输出，和一些元数据

> Bitcoin Core 将每个区块存储为磁盘上的不同文件，这样就不需要为了读取单一的块而将所有（或者部分）的块都加载到内存中，提高了性能

##### 在 `blocks` 中，key -> value 为：

| key | value |
| ---- | ---- |
| b | 32 字节的 `block hash`：块索引记录 |
| f | 4 字节的 `file number`：文件信息记录 |
| l | 4 字节的 `file number`： 最后使用的块文件号 |
| R |  1 字节的布尔值：是否在重建索引 |
| F | 1 字节的布尔值 flag `name length + flag name string`：可以打开或关闭的各种标志 |
| t | 32 字节的 `transaction hash`：交易索引记录 |

##### 在 `chainstate`，key -> value 为：

| key | value |
| ---- | ---- |
| c | 32 字节的 `transaction hash`：该交易的未花费交易输出记录 |
| B | 32 字节的 `block hash`：未花费交易的输出的块哈希值 |

[这里](https://en.bitcoin.it/wiki/Bitcoin_Core_0.11_(ch_2):_Data_Storage)可以查看 `Bitcoin Core` 关于存储的更多信息

> 目前还没有交易，所以只需要 `blocks bucket`，这边会将整个数据库存储为单个文件，而不是将区块存储在不同的文件中。 所以也不会需要文件编号（`file number`）相关的东西。
> 
> 最终，我们会用到的键值对有：

| key | value |
| ---- | ---- |
| b | 32 字节的 `block hash`：block 结构 |
| l | 链中最后一个块的 hash |

#### 流程

1. 打开一个数据库文件
2. 检查文件是否存储了一个区块链
3. 如果已经存储了一个区块链：
    1. 创建一个新的 `Blockchain` 实例
    2. 设置 `Blockchain` 实例的 `tip` 为数据库中存储的最后一个块的哈希
4. 如果没有区块链：
    1. 创建创世块
    2. 存储到数据库
    3. 将创世块哈希保存为最后一个块的哈希
    4. 创建一个新的 `Blockchain` 实例，初始时 `tip` 指向创世块
> tip: 末梢、尖端，这里 tip 表示存储的最后一个块的哈希

### 钱包

#### 地址
通过 Base58 将公钥转换成可读格式，包括以下步骤：

1. 使用 `RIPEMD16(SHA256(PubKey))` 对公钥进行两次哈希，生成 `pubKeyHash`
2. 将版本信息追加到 `pubKeyHash` 之前，生成 `versionedPayload`，此时 `versionedPayload = version + pubKeyHash`
3. 使用 `SHA256(SHA256(versionedPayload))` 进行两次哈希得到一个哈希值，取该值的前 `n` 个字节最终生成 `checksum`
4. 将 `checksum` 追加到 `versionedPayload` 之后，生成编码前的地址，此时 `address = version + pubKeyHash + checksum`
5. 最后使用 `Base58` 对 `version + pubKeyHash + checksum` 编码生成最终的地址

### Merkle Tree

![image](https://i.328888.xyz/2023/01/23/OK74z.md.png)

每个块都会有一个 Merkle 树，它从叶子节点（树的底部）开始，一个叶子节点就是一个交易哈希（比特币使用双 SHA256 哈希）。叶子节点的数量必须是双数，但是并非每个块都包含了双数的交易。因为，如果一个块里面的交易数为单数，那么就将最后一个叶子节点（也就是 Merkle 树的最后一个交易，不是区块的最后一笔交易）复制一份凑成双数。

从下往上，两两成对，连接两个节点哈希，将组合哈希作为新的哈希。新的哈希就成为新的树节点。重复该过程，直到仅有一个节点，也就是树根。根哈希然后就会当做是整个块交易的唯一标示，将它保存到区块头，然后用于工作量证明。

Merkle 树的好处就是一个节点可以在不下载整个块的情况下，验证是否包含某笔交易。并且这些只需要一个交易哈希，一个 Merkle 树根哈希和一个 Merkle 路径。

### 网络节点


在第一个终端窗口中将 `NODE_ID` 设置为 3000

```bash
$ export NODE_ID=3000
```
#### NODE 3000

创建一个钱包和一个新的区块链：

```bash
$ ./tchain-xxx createblockchain -address CENTREAL_NODE
```
（为了简洁起见，我会使用假地址。）

然后，会生成一个仅包含创世块的区块链。我们需要保存块，并在其他节点使用。创世块承担了一条链标识符的角色（在 Bitcoin Core 中，创世块是硬编码的）

```bash
$ cp blockchain_3000.db blockchain_genesis.db 
```

#### NODE 3001

接下来，打开一个新的终端窗口，将 node ID 设置为 3001。这会作为一个钱包节点。通过 `./tchain-xxx createwallet` 生成一些地址，我们把这些地址叫做  WALLET_1, WALLET_2, WALLET_3.

#### NODE 3000

向钱包地址发送一些币：

```bash
$ ./tchain-xxx send -from CENTREAL_NODE -to WALLET_1 -amount 10 -mine
$ ./tchain-xxx send -from CENTREAL_NODE -to WALLET_2 -amount 10 -mine
```

`-mine` 标志指的是块会立刻被同一节点挖出来。我们必须要有这个标志，因为初始状态时，网络中没有矿工节点。

启动节点：

```bash
$ ./tchain-xxx startnode
```

这个节点会持续运行，直到本文定义的场景结束。

#### NODE 3001

启动上面保存创世块节点的区块链：

```bash
$ cp blockchain_genesis.db blockchain_3001.db
```

运行节点：

```go
$ ./tchain-xxx startnode
```

它会从中心节点下载所有区块。为了检查一切正常，暂停节点运行并检查余额：

```bash
$ ./tchain-xxx getbalance -address WALLET_1
Balance of 'WALLET_1': 10

$ ./tchain-xxx getbalance -address WALLET_2
Balance of 'WALLET_2': 10
```

你还可以检查 `CENTRAL_NODE` 地址的余额，因为 node 3001 现在有它自己的区块链：

```bash
$ ./tchain-xxx getbalance -address CENTRAL_NODE
Balance of 'CENTRAL_NODE': 10
```

#### NODE 3002

打开一个新的终端窗口，将它的 ID 设置为 3002，然后生成一个钱包。这会是一个矿工节点。初始化区块链：

```bash
$ cp blockchain_genesis.db blockchain_3002.db
```

启动节点：

```bash
$ ./tchain-xxx startnode -miner MINER_WALLET
```

#### NODE 3001

发送一些币：

```bash
$ ./tchain-xxx send -from WALLET_1 -to WALLET_3 -amount 1
$ ./tchain-xxx send -from WALLET_2 -to WALLET_4 -amount 1
```

#### NODE 3002

迅速切换到矿工节点，你会看到挖出了一个新块！同时，检查中心节点的输出。

#### NODE 3001

切换到钱包节点并启动：

```bash
$ ./tchain-xxx startnode
```

它会下载最近挖出来的块！

暂停节点并检查余额：

```bash
$ ./tchain-xxx getbalance -address WALLET_1
Balance of 'WALLET_1': 9

$ ./tchain-xxx getbalance -address WALLET_2
Balance of 'WALLET_2': 9

$ ./tchain-xxx getbalance -address WALLET_3
Balance of 'WALLET_3': 1

$ ./tchain-xxx getbalance -address WALLET_4
Balance of 'WALLET_4': 1

$ ./tchain-xxx getbalance -address MINER_WALLET
Balance of 'MINER_WALLET': 10

## Build

在终端中执行
```bash
$ sudo ./build.sh
```
将会得到可在终端运行的可执行文件

|名称|操作系统|
| ---- | ---- |
| tchain-amd | Intel MacOS |
| tchain-arm | ARM MacOS |
| tchain-windows | Windows |

以 `tchain-amd` 为例，进入到可执行文件所在目录：

```bash
$ chmod 777 tchain-amd
```

创建钱包，生成一个地址，多次调用可以创建多个地址
```bash
$ ./tchain-amd createwallet
```

获取创建的所有的地址
```bash
$ ./tchain-amd listaddresses
```

生成创世块，向第一个用户发放Token：
```bash
$ ./tchain-amd createblockchain -address <YOUR_ADDRESS>
```

获取Token余额：
```bash
$ ./tchain-amd getbalance -address <YOUR_ADDRESS>
```

向其他人发送Token：
```bash
$ ./tchain-amd send -from <YOUR_ADDRESS> -to <YOUR_ADDRESS> -amount <AMOUNT>
```

打印区块链：
```bash
$ ./tchain-amd printchain
```
启动节点：
```bash
$ ./tchain-amd startnode
```

## 参考资料
- [1] [Mastering Bitcoin](https://github.com/bitcoinbook/bitcoinbook)
- [2] [UTXO 与账户余额模型](https://draveness.me/utxo-account-models/)
- [3] [Ivan Kuznetsov's Blog](https://jeiwan.net/)
