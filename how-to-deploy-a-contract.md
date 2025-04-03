# 在以太坊链（仅支持 Geth）上部署合约

## 快速开始

一个包含所有依赖的容器可以在 [这里](https://pan.baidu.com/s/1HwG3DNvNb32SxbQ1pyMwYQ) 找到。（密码：`hgvv`）

### 第一步：加载镜像并启动容器
```
docker load<contract_deployer.tar && docker run -i -t contractfuzzer/deployer
```

### 第二步：在容器中部署示例合约 `contract_deployer/contracts/`：

```
  contract_deployer/contracts
                      config
                      verified_contract_abis
                      verified_contract_bins
```
以下过程将尝试将 `Aeternis` 部署到私有链。

运行：
```
cd /ContractFuzzer && ./deployer_run.sh
```

### 第三步：检查 `Aeternis` 是否已部署。
查看文件 `/ContractFuzzer/contract_deployer/contracts/config/Aeternis.json`。

#### 部署前的配置：
```
"contracts": [
        {
            "home": "/ContractFuzzer/contract_deployer",
            "childhome": "/contracts",
            "from": "0x2b71cc952c8e3dfe97a696cf5c5b29f8a07de3d8",
            "gas": "50000000000",
            "name": "Aeternis",
            "param_Names": [
                "_owner"
            ],
            "param_Types": [
                "address"
            ],
            "param_Values": {
                "_owner": "0xed161fa9adad3ba4d30c829034c4745ef443e0d9"
            },
            "values": [
                "0xed161fa9adad3ba4d30c829034c4745ef443e0d9"
            ],
            "payable": false,
            "value": "1000000000"
        }
```

#### 部署成功后：
如果部署成功，`address` 字段将被添加，并设置为 `Aeternis` 的私有链地址。
```
"contracts": [
        {
            "home": "/ContractFuzzer/contract_deployer",
            "childhome": "/contracts",
            "from": "0x2b71cc952c8e3dfe97a696cf5c5b29f8a07de3d8",
            "gas": "50000000000",
            "name": "Aeternis",
            "param_Names": [
                "_owner"
            ],
            "param_Types": [
                "address"
            ],
            "param_Values": {
                "_owner": "0xed161fa9adad3ba4d30c829034c4745ef443e0d9"
            },
            "values": [
                "0xed161fa9adad3ba4d30c829034c4745ef443e0d9"
            ],
            "payable": false,
            "value": "1000000000",
            "address": "0xbcf6fb693173f2a6c7c837a31717c403b496ccae"
        }
```

---

## 部署以太坊合约

### 前置条件

1. 提供合约的 ABI 定义文件。
2. 提供合约的 BIN 文件。
3. 提供合约的配置文件。

你可以从 Docker 中的现有合约学习这些文件的格式：
```
  contract_deployer/contracts
                        config
                        verified_contract_abis
                        verified_contract_bins
```
- `config` 目录包含合约部署的配置文件。你可以根据需要部署的合约参数复制并修改这些文件。
- `verified_contract_abis` 目录包含从 Etherscan 下载的 ABI 文件。
- `verified_contract_bins` 目录包含从 Etherscan 下载的 BIN 文件（直接保存合约创建代码到 BIN 文件中）。

运行以下命令：
```
docker run -it -v YourEthereumPrivateChainPath:/ContractFuzzer/Ethereum -v your_contracts_to_deploy:/ContractFuzzer/contract_deployer/contracts  -e "ContractFuzzer=/contractFuzzer/contract_deployer"  ContractFuzzer/deployer:latest
```

然后运行：
```
cd /ContractFuzzer && ./deployer_run.sh
```

最后，你可以在文件 `/ContractFuzzer/contract_deployer/contracts/config/xxx.json` 中找到合约的 `address`！

---

## 注意事项

请注意，合约的部署可以在 Docker 容器中完成，也可以在本地机器上完成，只要你准备好了配置文件、BIN 文件和 ABI 文件。在本地机器上，启动 Geth 客户端后，你可以运行 `./deployer_run.sh` 脚本来部署智能合约。

本说明正在更新中，部署成功可能需要一些额外的努力。

此外，如果你能够通过自己的方式在私有链（基于我们提供的 `Ethereum` 基础链）上部署合约，那就更好了。
