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

*被交易消费的UTXO被称为交易输入，由交易创建的UTXO被称为交易输出。每一笔交易的来源(输入)都来自于上一笔交易的输出。*



所以没有人能真正拥有比特币，你看到的地址账户里的比特币，是”钱包“帮你在UTXO模型里算出来的，其实它的“价值”还躺在上一笔交易里呢！

#### Coinbase
对于输出与输入链来说，有一种特殊的交易，称为“币基交易”，它是每个区块的第一笔交易。
这种交易是由获胜的矿工放置的，给这个矿工创建了新的可花费比特币，它是挖矿的奖励。
这个特殊的币基交易不消费UTXO，它有一个特殊类型的输入，称为“币基”。
这就是挖矿过程中创建比特币货币供应的方式。
> 输入和输出，哪个先产生？先有鸡还是先有蛋？
> 严格来讲，先产生输出，因为币基交易（创造新比特币）没有输入，它是无中生有。



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

生成创世块，向第一个用户发放Token：
```bash
$ ./tchain-amd createblockchain -address <YOUR_NAME>
```

获取Token余额：
```bash
$ ./tchain-amd getbalance -address <YOUR_NAME>
```

向其他人发送Token：
```bash
$ ./tchain-amd send -from <YOUR_NAME> -to <OTHER_NAME> -amount <AMOUNT>
```

打印区块链：
```bash
$ ./tchain-amd printchain
```

## 参考资料
- [1] [Mastering Bitcoin](https://github.com/bitcoinbook/bitcoinbook)
- [2] [UTXO 与账户余额模型](https://draveness.me/utxo-account-models/)
- [3] [Ivan Kuznetsov's Blog](https://jeiwan.net/)