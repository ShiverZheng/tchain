# TChain

## 基本原型

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