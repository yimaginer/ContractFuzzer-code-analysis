// 导入所需模块
import * as fs from "fs"; // 文件系统模块，用于文件操作
export const dotenv = require("dotenv"); // 加载环境变量配置
export const assert = require("assert"); // 断言模块，用于测试和验证
export const Promise = require("bluebird"); // Promise 库，用于异步操作
export const Web3 = require("web3"); // Web3.js，用于与以太坊节点交互
export const solc = require("solc"); // Solidity 编译器
export const truffle_Contract = require("truffle-contract"); // Truffle 合约库，用于部署和交互智能合约

// 加载环境变量
dotenv.config();

// 定义文件写入的 Promise 包装
const writeFilePs = Promise.promisify(fs.writeFile);

// 从环境变量中获取以太坊客户端的 RPC 地址
const HttpRpcAddr = process.env.GethHttpRpcAddr;
console.log(HttpRpcAddr);

// 创建 Web3 提供程序
export const Provider = new Web3.providers.HttpProvider(HttpRpcAddr);
export const web3 = new Web3(Provider);

// 获取以太坊账户信息
const accounts = web3.eth.accounts; // 所有账户
const defaultAccount = web3.eth.accounts[0]; // 默认账户
const defaultGas = 800000; // 默认 Gas 限额
const defaultValue = 100000; // 默认交易附带的以太币

// 定义默认交易参数（带以太币）
export const defaultAmountParamsWithValue = {
    from: defaultAccount,
    value: defaultValue,
    gas: defaultGas
};

// 定义默认交易参数（不带以太币）
export const defaultAmountParamsWithoutValue = {
    from: defaultAccount,
    gas: defaultGas
};

// 合约工具类
export class ContractUtils {
    Password = "123456"; // 默认密码
    GasPrice = 8000000; // 默认 Gas 价格

    constructor() {
        // 构造函数
    }

    // 异步读取文件内容
    async _read(file) {
        let readFile = Promise.promisify(fs.readFile);
        let res = await readFile(file);
        return res.toString(); // 返回文件内容的字符串
    }

    // 部署带参数的合约
    async DeployWithParams(contract, iParamsArr, aParams) {
        aParams.value = web3.toBigNumber(aParams.value); // 转换 value 为 BigNumber
        aParams.gas = web3.toBigNumber(aParams.gas); // 转换 gas 为 BigNumber
        const owner = this.defaultAccount; // 默认账户
        let contract_name_ = contract.name; // 合约名称
        let abi_ = contract.abi; // 合约 ABI
        let code_ = contract.bin; // 合约字节码

        // 使用 Truffle 合约库创建合约实例
        let MyContract = truffle_Contract({
            contract_name: contract_name_,
            abi: abi_,
            unlinked_binary: code_,
            default_network: 1900 // 默认网络 ID
        });
        MyContract.setProvider(Provider); // 设置 Web3 提供程序
        let instance = await MyContract.new(...iParamsArr, aParams); // 部署合约
        return instance; // 返回合约实例
    }

    // 部署不带参数的合约
    async DeployWithoutParams(contract, aParams) {
        aParams.value = web3.toBigNumber(aParams.value); // 转换 value 为 BigNumber
        aParams.gas = web3.toBigNumber(aParams.gas); // 转换 gas 为 BigNumber
        const owner = this.defaultAccount; // 默认账户
        let contract_name_ = contract.name; // 合约名称
        let abi_ = contract.abi; // 合约 ABI
        let code_ = contract.bin; // 合约字节码

        // 使用 Truffle 合约库创建合约实例
        let MyContract = truffle_Contract({
            contract_name: contract_name_,
            abi: abi_,
            network_id: 1900, // 网络 ID
            unlinked_binary: code_,
            default_network: 1900 // 默认网络 ID
        });
        MyContract.setProvider(Provider); // 设置 Web3 提供程序
        let instance = await MyContract.new(aParams); // 部署合约
        return instance; // 返回合约实例
    }

    // 获取已部署的合约实例
    async _instance(contract) {
        let name_ = contract.name; // 合约名称
        let abi_ = JSON.parse(unescape(contract.abi)); // 解码并解析 ABI
        let code_ = unescape(contract.code); // 解码字节码
        let address_ = contract.address; // 合约地址

        // 使用 Truffle 合约库创建合约实例
        let MyContract = truffle_Contract({
            contract_name: name_,
            abi: abi_,
            unlinked_binary: code_,
            network_id: 1900, // 网络 ID
            address: address_, // 合约地址
            default_network: 1900 // 默认网络 ID
        });
        MyContract.setProvider(Provider); // 设置 Web3 提供程序
        let instance = await MyContract.deployed(); // 获取已部署的合约实例
        return instance; // 返回合约实例
    }

    // 测试函数，打印消息
    show() {
        console.log("hello ContractUtils");
    }
}

// 使用示例（注释掉的代码）
// const utils = new ContractUtils();
// utils.test_promise();
// utils._read("./Contracts/SendBalance_V3.sol");