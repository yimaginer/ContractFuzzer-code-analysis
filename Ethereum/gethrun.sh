# 启动 Geth 节点的命令，包含多个选项，用于配置以太坊客户端的行为
geth \
    --fast \ # 启用快速同步模式，仅下载区块头和状态数据，而不是完整区块数据
    --identity "TestNode2" \ # 设置节点的标识名称为 "TestNode2"
    --rpc \ # 启用 RPC 服务，允许外部客户端通过 HTTP 访问节点
    --rpcaddr "0.0.0.0" \ # 将 RPC 服务绑定到所有网络接口（0.0.0.0）,允许所有ip远程访问
    --rpcport "8545" \ # 设置 RPC 服务的端口号为 8545
    --rpccorsdomain "*" \ # 允许所有域名通过 CORS 访问 RPC 服务（不推荐用于生产环境）
    --port "30303" \ # 设置节点的 P2P 网络通信端口为 30303
    --nodiscover \ # 禁用节点发现功能，节点不会自动发现其他节点
    --rpcapi "db,eth,net,web3,miner,net,personal,net,txpool,admin" \ # 启用的 RPC API 模块，包括数据库、以太坊、网络、Web3、挖矿、账户管理、交易池和管理模块
    --networkid 1900 \ # 设置网络 ID 为 1900，用于区分不同的以太坊网络
    --datadir /home/liuye/Ethereum \ # 指定数据目录为 /home/liuye/Ethereum，用于存储区块链数据
    --nat "any" \ # 设置 NAT 类型为 "any"，自动检测 NAT 配置
    --targetgaslimit "9000000000000" \ # 设置目标 Gas 限额为 9000000000000
    --unlock 0 \ # 解锁第一个账户（索引为 0），用于发送交易或挖矿
    --password "pwd.txt" \ # 指定包含账户密码的文件 pwd.txt，用于解锁账户
    --mine # 启用挖矿模式，节点将开始挖矿