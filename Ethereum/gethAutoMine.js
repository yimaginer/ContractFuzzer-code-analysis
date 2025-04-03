// Geth 自动挖矿脚本
// 功能：
// 1. 检测是否有交易提交到交易池中。
// 2. 自动启动挖矿以将交易打包到区块链中。

// 何时调用: 在私有链或测试环境中，节点启动后需要自动挖矿时调用。
// 谁调用: 开发者、测试人员或自动化工具。
// 如何调用: 通过 Geth 的 loadScript 或 --preload 参数加载并运行脚本

// 获取主账户（默认账户）
var primary = eth.accounts[0];

// 解锁主账户，设置密码和解锁时间（单位：秒）
personal.unlockAccount(primary, "123456", 200 * 60 * 60); // 解锁 200 小时
personal.unlockAccount(eth.accounts[1], "123456", 200 * 60 * 60); // 解锁第二个账户

// 设置挖矿奖励账户（etherbase）
miner.setEtherbase(primary);

// 定义检查间隔时间（单位：秒）
var INTEVAL = 60; // 在私有链中设置为 60 秒

// 无限循环，持续检测交易池状态
while (true) {
    // 获取交易池状态
    var status = txpool.status;
	// txpool 是 Geth 提供的全局对象，用于管理和查询交易池。
	// 它的主要功能是监控交易池的状态（status）、查看交易详情（content）和快速检查交易（inspect）。

    // 打印交易池状态（可选）
    // console.log("pending:" + status.pending); // 打印待处理交易数量
    // console.log("queued:" + status.queued); // 打印排队交易数量

    // 如果交易池中有待处理或排队的交易
    if (status.pending != 0 || status.queued != 0) {
        console.log("mine......"); // 打印挖矿开始信息
        miner.start(); // 启动挖矿
        admin.sleepBlocks(3); // 等待 3 个区块被挖出
        miner.stop(); // 停止挖矿
        console.log("finished!"); // 打印挖矿完成信息
    }

    // 等待一段时间后再次检查交易池状态
    admin.sleep(INTEVAL / 10); // 休眠时间为间隔的 1/10
    // admin.sleep(1); // 可选：每秒检查一次
}
