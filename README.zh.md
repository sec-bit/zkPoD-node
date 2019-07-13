
# zkPoD: A decentralized system for data exchange

*You can also read this in English [here](README.md).*

## 概览

zkPoD 是一个用于「不可信双方」之间进行数字商品（数据）交易的去中心化平台。无需依赖任何可信第三方（中间商），即可实现「一手交钱一手交货」。zkPoD 使用区块链（例如以太坊）作为「无需信任」的第三方来确保公平性，任何一方都无法在交易中作弊获利。而且，zkPoD 注重用户隐私，向区块链矿工和其他各方保护用户意图。

任何卖家均可发布数据用于出售，覆盖以下两个场景：

- 数据下载：卖家从数据卖家手中付费购买并下载一份数据文件。zkPoD 支持数据片段交易，即买家可在一笔批量交易中下载任意指定的数据块。
- 数据查询：zkPoD 支持结构化数据。例如，卖家将数据以表格的形式编排，并可指定多列为索引字段，用户则可使用一个或多个关键词进行付费查询，进而获取匹配的记录。zkPoD 可保证查询结果是可信赖的，即：(i) 如果数据卖家返回了 n 组记录，那么在该表中不可能存在更多匹配查询关键词的记录；(ii) 返回的 n 组记录在该表中一定确切存在，不会存在返回结果伪造的问题。

zkPoD 重点解决以下三个主要问题：

- 买家最终得到的数据正是其付费前想要的数据
- 数据必须在卖家完成支付的同时完成交付
- 数据在完成交易前不会发生任何泄露

我们设计了名为 PoD (Proof of Delivery) 的密码学协议来尝试解决以上问题，确保数据买卖双方间的交易公平性。PoD 协议是零知识和可证明安全的（证明工作正在进行）。想要了解更多信息请查看技术白皮书。

zkPoD 实用且高效，在普通 PC 上即能支持大小为 10 GB 的文件，并且理论上支持数 TB 级别的数据交易。详情可查看性能评估小节。

[![asciicast-gif](img/demo.min.gif)](https://asciinema.org/a/251240?autoplay=1&speed=2.71828182846)

## 亮点

- 完全去中心化：zkPoD 利用以太坊上的智能合约作为去信任的第三方，并且理论上可部署至任意支持基本智能合约功能的区块链。数据交易链上消耗（Gas）适中，数据交易容量上限可达数 TB。
- 支持原子交换：zkPoD 支持原子交换（效果如 [ZKCP](https://en.bitcoin.it/wiki/Zero_Knowledge_Contingent_Payment)）。
- 支持大容量数据交易：zkPoD 支持在一笔交易中完成大容量数据的验证。参见性能评估小节。
- 支持关键词数据查询：zkPoD 支持付费查询。买家可发起包含一个或多个关键词的付费查询请求，来定位感兴趣的数据记录。
- 隐私保护：买家的购买请求在某些场景下是十分敏感的隐私信息，zkPoD 允许买家通过添加一些无关的请求，来混淆自己的真实意图。与此同时，卖家收到请求后并不清楚对方的真正目标，必须对所有请求逐一作出回应，但卖家知道所有回应中只有一个是对买家可见的，因为买家毕竟只为其中一个购买请求进行了付费。
- 支持验货：zkPoD 原生支持任意颗粒度的验货。买家可先随机抽样购买任意位置、任意数
  量的数据，进行验货，确认数据无误后再进行大批量购买。zkPoD 对验货次数不做任何限
  制，并且可保证每次验货数据（包括最后大批量购买）都来自同一数据集。

## 工作流程和原理

我们通过一个简化版的 PoD 协议来简述 zkPoD 的交易流程。

#### 数据初始化

下文中，Alice 代表卖家，Bob 代表卖家。

数据在出售前需要经过预处理。Alice 提前计算得到待出售数据文件的一组“验证器”（authenticator）和“验证器”的 Merkle root。“验证器”用于数据内容和来源的验证（即使数据处于被加密）。

zkPoD 支持两种模式：plain 模式和 table 模式。

- plain（普通文件）
- table（结构化文件，目前支持 CSV 格式）

对于制表数据，文件的每一行都对应一条有着固定列的记录。买家可以发起包含特定关键词的查询请求。注意，卖家在数据初始化时必须预先指定若干个列作为索引字段，以支持关键词查询。

#### 数据交易

zkPoD 支持两种交易模式的数据传输（交易）。

- Atomic-swap mode

1. Bob 发现感兴趣的数据，下载公开的数据标签，发起购买请求
2. Alice 返回给 Bob 对应的加密数据（使用一次性随机密钥进行加密）
3. Bob 通过 ZKP 校验加密数据和数据标签间的关系
4. Bob 认可返回的数据，向合约（区块链）提交收据
5. Alice 检查收据，然后向合约披露解密数据需要的 key
6. 合约（区块链）根据 receipt 中的参数验证 key 是否正确，输出“接受”或“拒绝”

- Complaint mode (受 Fairswap 启发)

1. Bob 发现感兴趣的数据，下载公开的数据标签，发起购买请求
2. Alice 返回给 Bob 对应的加密数据（使用一次性随机密钥进行加密）
3. Bob 通过 ZKP 校验加密数据和数据标签间的关系
4. Bob 认可返回的数据，向合约（区块链）提交收据
5. Alice 检查收据，然后向合约披露解密数据需要的 key
6. Bob 通过 key 解密数据，如果发现 Alice 作弊，则向合约（区块链）提交作弊证明

### 背后原理

为了交易双方的公平性和安全性，zkPoD 协议确保了以下要点：

{1}. 合约（区块链）无法获知交易数据和加密数据的任何内容
{2}. Bob 必须提交正确的 receipt 来获得 key
{3}. Bob 必须在获得 key 前进行支付
{4}. Bob 无法从加密的数据中获得任何信息
{5}. Alice 不能披露假的 key，这种情况会被合约中的校验算法排除
{6}. Alice 不能将数据替换成无关的垃圾数据给 Bob，这种情况无法通过加密数据与数据标签的验证步骤

为了确保第 **{1, 4, 6}** 点，我们使用了基于 Pedersen commitments（具备加法同态性质）的零知识证明，结合一次性密码本加密，从而允许买家无需借助他人帮助来完成数据的验证。zkPoD 系统中智能合约被用于以一种透明、可预测、可形式化验证的方式完成加密货币与解密密钥的互换交易。

我们使用「可验证随机函数」(Verifiable Random Function, VRF) 来支持关键词查询。目前，zkPoD 暂时只支持关键词精确匹配。zkPoD 还采用「不经意传输」(Oblivious Transfer, OT) 来支持隐私保护的查询。


## 编译与使用

### Build

*WIP: A building script for all of these steps*

#### 1. Build zkPoD-lib

请先[参考此处](https://github.com/sec-bit/zkPoD-lib#dependencies)安装 zkPoD-lib 的相关依赖。

```shell
# Download zkPoD-lib code
mkdir zkPoD && cd zkPoD
git clone https://github.com/sec-bit/zkPoD-lib.git

# Pull libsnark submodule
cd zkPoD-lib
git submodule init && git submodule update
cd depends/libsnark
git submodule init && git submodule update

# Build libsnark
mkdir build && cd build
# - For Ubuntu
cmake -DCMAKE_INSTALL_PREFIX=../../install -DMULTICORE=ON -DWITH_PROCPS=OFF -DWITH_SUPERCOP=OFF -DCURVE=MCL_BN128 ..
# - Or for macOS (see https://github.com/scipr-lab/libsnark/issues/99#issuecomment-367677834)
CPPFLAGS=-I/usr/local/opt/openssl/include LDFLAGS=-L/usr/local/opt/openssl/lib PKG_CONFIG_PATH=/usr/local/opt/openssl/lib/pkgconfig cmake -DCMAKE_INSTALL_PREFIX=../../install -DMULTICORE=OFF -DWITH_PROCPS=OFF -DWITH_SUPERCOP=OFF -DCURVE=MCL_BN128 ..
make && make install

# Build zkPoD-lib
cd ../../..
make

# These files should be generated after successful build.
# pod_setup/pod_setup
# pod_publish/pod_publish
# pod_core/libpod_core.so
# pod_core/pod_core

cd pod_go
export GO111MODULE=on
make test
```

#### 2. Build zkPoD-node

```shell
cd zkPoD
git clone https://github.com/sec-bit/zkPoD-node.git
cd zkPoD-node
export GO111MODULE=on
make
```

### Have Fun

#### 1. Setup

我们需要 [trusted setup](https://z.cash/technology/paramgen/) 来生成 zkPoD zkSNARK 的公共参数。

为了方便且仅处于测试目的，我们可以直接从 [zkPoD-params](https://github.com/sec-bit/zkPoD-params) 仓库进行下载。

```shell
cd zkPoD-node
mkdir -p key/zksnark_key
cd key/zksnark_key
# Download zkSNARK pub params, see https://github.com/sec-bit/zkPoD-params
wget https://raw.githubusercontent.com/sec-bit/zkPoD-params/master/zksnark_key/atomic_swap_vc.pk
wget https://raw.githubusercontent.com/sec-bit/zkPoD-params/master/zksnark_key/atomic_swap_vc.vk
```

#### 2. Run node

```shell
cd zkPoD-node
make run
# A config file named basic.json is generated on local
```
> Examples: [`basic.json`](examples/basic.json) - zkPoD-node 的基础配置文件

提示：

在 Linux 上运行 `zkPoD-node` 时，应该为 `libpod_core` 指定 `LD_LIBRARY_PATH`。在 macOS 上则应该指定 `DYLD_LIBRARY_PATH`。可以参考 `Makefile` 中的例子，为了方便起见，可将 `LD_LIBRARY_PATH` 设置为环境变量。

```shell
# On Linux
export LD_LIBRARY_PATH=<YOUR_PATH_TO_LIBPOD_CORE>

# Or on macOS
export DYLD_LIBRARY_PATH=<YOUR_PATH_TO_LIBPOD_CORE>
```

#### 3. 保存 keystore 文件，获取 ETH

- https://faucet.ropsten.be/
- https://faucet.metamask.io/

提示：zkPoD-node 首次启动后会自动创建一个全新的 Ethereum 地址，从终端日志中或 keystore 文件中可以读取。用户应该安全保管自己的 keystore 文件。由于该地址需要与以太坊合约发生交互，因此地址中必须留有一定的余额。测试阶段（Ropsten 测试网），你可以从上列 Ethereum faucet 网站获得免费的测试网络 ETH。

#### 4. 卖家: 数据初始化，发布数据

打开一个新的终端

```shell
cd zkPoD-node
cp ../zkPoD-lib/pod_publish/pod_publish .

wget -O test.txt https://www.gutenberg.org/files/11/11-0.txt

./zkPoD-node -o initdata -init init.json
# You should get the sigma_mkl_root from logs
# export sigma_mkl_root=<YOUR_SIGMA_MKL_ROOT>
./zkPoD-node -o publish -mkl $sigma_mkl_root -eth 200
# You should get the publish transaction hash
```
> Examples: [init.json](examples/init.json) - 使用该配置文件来描述用于出售的数据

提示：你可以使用相同的地址进行测试，完成出售和购买。你还可以长期运行 zkPoD-node 节点程序，在[社区](https://discord.gg/tfUH886)频道内公布待出售数据的信息，与其他玩家一起完成公平交易测试。

如果你想出售数据，你应该让其他玩家知道以下信息：

```
- 你节点程序的 IP 地址
- 你的 ETH 地址
- 用于出售的数据的 sigma_mkl_root
- 数据的描述信息
- 数据初始化后生成的 bulletin 文件
- 数据初始化后生成的 public 信息
```

待售数据初始化完成后，卖家可以从该路径 `zkPoD-node/seller/publish/$sigma_mkl_root/` 获得数据的 `bulletin` 和 `public` 信息。

```
├── bulletin
├── extra.json
├── private
│   ├── matrix
│   └── original
├── public
│   ├── sigma
│   └── sigma_mkl_tree
└── test.txt
```

#### 5. 买家: 向合约抵押 ETH

作为买家，如果你在社区中发现了感兴趣的数据，想从卖家手上进行购买。你需要预先向 zkPoD 交易合约抵押 ETH 用于后续交易。请放心，在得到你想要的数据前，你的 ETH 仍然还是你的。

```shell
./zkPoD-node -o deposit -eth 20000 -addr $SELLER_ETH_ADDR #卖家地址
# You should get the deposit transaction hash
```

#### 6. 买家: 进行购买

买家将向卖家发起购买请求。为了方便，你可以在配置文件中填入卖家和数据的一些基本信息。

```shell
# For test, you could simply copy public info of data from seller folder to project root path.
# cp seller/publish/$sigma_mkl_root/bulletin .
# cp -r seller/publish/$sigma_mkl_root/public .
./zkPoD-node -o purchase -c config.json
# You should get the decrypted data in buyer/transaction/<session_id> folder
```
> Examples: [config.json](examples/config.json) - 使用该文件描述你想要购买的数据。

提示：
1. Atomic-swap 模式目前在以太坊网络上仅支持最大 340 KiB 大小的数据交易。

2. 如果选择了 complaint 模式，zkPoD-node 节点程序会自动向合约发起申诉，并提供卖家的作弊证明。因此，不诚实的卖家无法通过作弊而获利。

TODO: 还有更多好玩的功能，后续会添加更多的使用方法例子介绍，如对表格数据进行普通查询和私密查询。

## zkPoD 项目结构

![](img/overview.svg)

- [zkPoD-node](https://github.com/sec-bit/zkPoD-node) 节点应用程序（Golang），供买卖双方使用，负责处理通信、合约查询与调用、数据传输以及其他 zkPoD 的协议交互。
- [zkPoD-lib](https://github.com/sec-bit/zkPoD-lib) zkPoD 底层核心库（C++），同时提供 Golang binding。
- [zkPoD-contract](https://github.com/sec-bit/zkPoD-contract) 智能合约（Solidity），实现 zkPoD 数据去中心化交易功能。

## 性能评估

#### 测试环境

- OS: Ubuntu 16.04.6 LTS x86_64
- CPU Model: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
- CPU Thread Count: 12
- Memory: 32605840 kB

#### 基本信息

We present three variant protocols, PoD-AS, PoD-AS* and PoD-CR, used for different purposes.

为了满足不同应用场景，我们设计了三个变种协议 PoD-AS、PoD-AS* 和 PoD-CR。

|  Protocol  | Throughput |   Communication   |   Gas Cost (Ethereum)   | Data/Tx (Ethereum) |
| :----: | :----------------: | :---------------------: | :---------------------: | :---------------------: |
| PoD-CR |        3.39 MiB/s       |        $O(2n)$        | $O(\log{}n)$ |         < 100 TiB         |
| PoD-AS |        3.91 MiB/s       |    $O(2n)$    |    $O(n)$    |        < 350 KiB        |
| PoD-AS* |    35 KiB/s    |    $O(2n)$    |    $O(1)$    |        Unlimited        |

PoD-AS 协议数据传输速度最快，链上计算复杂度为 O(n)。此协议十分适合高 TPS、低链上运算成本的许可链。

PoD-AS* 协议使用 zkSNARKs 来降低链上计算量至 O(1)，但是链下传输速度较慢（计算量大）。

PoD-CR 协议支持较快的数据传输速度和较低的链上计算量。

#### 测试结果

- Data size: 1024 MiB
- File type: plain
- s: 64
- omp_thread_num: 12

|      Protocol      | Prover (s) | Verifier (s) | Decrypt (s) | Communication Traffic (MiB) | Gas Cost |
| :------------: | :--------: | :----------: | :---------: | :-------------------------: | :------: |
| PoD-CR |    124     |     119      |     82      |            2215             | 159,072  |
|  PoD-AS   |    130     |     131      |    4.187    |            2215             |   `*`    |
|  PoD-AS*   |    34540     |     344      |    498    |            2226             |   183,485   |

`*` PoD-AS 协议在以太坊网络上暂时不支持交易 1 GiB 大小的文件.

#### 以太坊网络上的 Gas 消耗

PoD-CR Protocol            |  PoD-AS Protocol      |  PoD-AS* Protocol
:-------------------------:|:-------------------------:|:-------------------------:
![](img/Gas-Cost-vs-Data-Size-Batch1.svg)  | ![](img/Gas-Cost-vs-Data-Size-Batch2.svg) | ![](img/Gas-Cost-vs-Data-Size-Batch3.svg) 

## 想要了解更多？

+ 白皮书：zkPoD 系统的整体介绍
+ 技术白皮书： zkPoD 的详细技术细节
+ 社区: 欢迎加入我们的 [*Discord*](https://discord.gg/tfUH886) 参与讨论，关注 [*Twitter*](https://twitter.com/SECBIT_IO) 获取最新动态。

## 其他相关项目

+ Fairswap:  https://github.com/lEthDev/FairSwap
+ ZKCP: https://en.bitcoin.it/wiki/Zero_Knowledge_Contingent_Payment
+ Paypub: https://github.com/unsystem/paypub
