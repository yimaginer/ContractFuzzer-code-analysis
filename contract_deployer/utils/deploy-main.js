#! /local/bin/bnode

import * as fs from "fs";
import {
    ContractUtils,
    Contract_DIR,
    web3,
    defaultAmountParamsWithValue,
    defaultAmountParamsWithoutValue,
    dotenv,
} from "../utils/ContractUtils.js";
import {
    defaultAccount
} from "../utils/Account.js";

dotenv.config(); // 加载环境变量配置

// 定义目录和文件后缀
const BIN_DIR = process.env.BIN_SUB_DIR; // 合约二进制文件目录
const BIN_SUFFIX = process.env.BIN_SUFFIX; // 合约二进制文件后缀
const ABI_DIR = process.env.ABI_SUB_DIR; // 合约 ABI 文件目录
const ABI_SUFFIX = process.env.ABI_SUFFIX; // 合约 ABI 文件后缀
const CONFIG_PATH = process.env.CONFIG_PATH; // 合约配置文件路径

const util = new ContractUtils(); // 合约工具类实例

// 文件操作相关函数
const fread = fs.readFileSync;
const fopen = fs.openSync;
const fclose = fs.closeSync;
const fwrite = fs.writeFileSync;
const listdir = fs.readdirSync;
const print = console.log;
const error = console.error;
const BalanceOf = web3.eth.getBalance;

// 从 JSON 文件中读取数据
function deJSON(file) {
    let f = fopen(file, "r");
    let data = fread(f, {
        encoding: "utf8"
    });
    let js = JSON.parse(data);
    fclose(f);
    return js;
}

// 将 JSON 数据写入文件
function writeJSONtoFile(file, jsData) {
    let f = fopen(file, "w");
    fwrite(f, jsData, {
        encoding: "utf8"
    });
    fclose(f);
}

// 读取合约的二进制文件
function deBin(workplace, name) {
    let file = workplace + "/" + BIN_DIR + "/" + name + BIN_SUFFIX;
    let f = fopen(file, "r");
    let data = fread(f, {
        encoding: "utf8"
    });
    fclose(f);
    data = "0x" + data; // 添加前缀 "0x"
    return data;
}

// 读取合约的 ABI 文件
function deAbi(workplace, name) {
    let file = workplace + "/" + ABI_DIR + "/" + name + ABI_SUFFIX;
    let f = fopen(file, "r");
    let data = fread(f, {
        encoding: "utf8"
    });
    fclose(f);
    data = JSON.parse(data);
    return data;
}

// 获取 JSON 对象的长度
function getJsonObjLength(jsonObj) {
    var Length = 0;
    for (var item in jsonObj) {
        Length++;
    }
    return Length;
}

// 部署单个合约
async function deployAcontract(workplace, contract) {
    let name = contract.name; // 合约名称
    let param_Values = contract.param_Values; // 合约参数值
    if (param_Values == "none") {
        param_Values = {};
    }
    let from = contract.from; // 部署者地址
    let gas = contract.gas; // Gas 限额
    let value = contract.value; // 交易附带的以太币
    let payable = contract.payable; // 是否为可支付合约
    let values = contract.values; // 参数值数组
    let abi = deAbi(workplace, name); // 获取 ABI
    let bin = deBin(workplace, name); // 获取二进制代码
    contract = {
        name: name,
        abi: abi,
        bin: bin
    };
    let aParams;
    // 设置部署参数
    if (payable == true) {
        aParams = {
            from: from,
            value: value,
            gas: gas
        };
    } else {
        aParams = {
            from: from,
            gas: gas
        };
    }

    let len = getJsonObjLength(param_Values); // 检查参数长度
    if (len == 0) {
        // 无参数部署
        return await util.DeployWithoutParams(contract, aParams);
    } else {
        // 带参数部署
        let paramArr = values;
        return await util.DeployWithParams(contract, paramArr, aParams);
    }
}

// 部署多个合约
async function deploy(js) {
    let home = js.home; // 合约主目录
    let value = js.value; // 默认交易附带的以太币
    let gas = js.gas; // 默认 Gas 限额
    let contracts = js.contracts; // 合约列表
    let from = js.from; // 部署者地址
    if (from == undefined || from == "none") {
        from = web3.eth.accounts[0]; // 使用默认账户
    }
    for (let index in contracts) {
        let contract = contracts[index];
        if (contract.deployed != undefined && contract.deployed == 1)
            continue; // 跳过已部署的合约
        let workplace;
        if (contract.home == undefined || contract.home == "none")
            contract.home = home;
        if (contract.childhome == undefined || contract.childhome == "none")
            workplace = contract.home;
        else
            workplace = contract.home + contract.childhome;
        if (contract.gas == undefined || contract.gas == "none")
            contract.gas = gas;
        if (contract.value == undefined || contract.value == "none")
            contract.value = value;
        if (contract.from == undefined || contract.from == "none") {
            contract.from = from;
        }
        try {
            // 部署合约
            let instance = await deployAcontract(workplace, contract);
            contract.deployed = 1; // 标记为已部署
            contract["address"] = instance.address; // 保存合约地址
        } catch (err) {
            print("部署错误");
            print("合约名称:", contract.name);
            print(err.toString().split("\n")[0]);
            contract.deployed = 0; // 标记为部署失败
        }
    }
    console.log(js);
    writeJSONtoFile(js.file_path, JSON.stringify(js)); // 更新配置文件
    return JSON.stringify(js);
}

// 批量部署合约
function deployBatch(js_arr) {
    for (let js of js_arr) {
        deploy(js);
    }
}

// 获取配置文件列表
function getConfigs(dir) {
    let items = listdir(dir);
    let files = [];
    let index = 0;
    for (let i = 0; i < items.length; i++) {
        let item = items[i];
        files[index++] = dir + "/" + item;
    }
    return files;
}

// 部署流程:
// 读取配置文件。
// 解析配置文件中的合约信息。
// 调用 deployAcontract 部署合约。
// 将部署结果（如合约地址）写回配置文件。

// 执行外部配置文件中的部署
function executeOuterConfigs() {
    let dir = CONFIG_PATH; // 配置文件目录
    let configs = getConfigs(dir); // 获取配置文件列表
    let js_arr = [];
    const batchMaxSize = 20; // 每批最大部署数量
    const batchInterval = 30 * 1000; // 每批部署间隔时间（毫秒）
    let times = 0;
    for (let config of configs) {
        let js = deJSON(config); // 读取配置文件
        js.file_path = config;
        js_arr.push(js);
        if (js_arr.length > batchMaxSize) {
            setTimeout(deployBatch, times * batchInterval, js_arr); // 延迟批量部署
            times++;
            js_arr = [];
        }
    }
    console.log(js_arr);
    setTimeout(deployBatch, times * batchInterval, js_arr); // 部署剩余的合约
}

// 主函数
function main() {
    executeOuterConfigs(); // 执行外部配置文件中的部署
}

// 判断是否为主模块
let __name__ = "__main__";
if (__name__ == "__main__") {
    main();
}
