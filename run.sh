#!/bin/sh
# 该脚本用于启动 ContractFuzzer 的完整测试流程，包括 Geth 节点、测试器和模糊测试工具。

# 获取当前工作目录
DIR=${PWD}

# 打印传入的参数数量
echo $#

# 检查参数数量是否为 2
if [ $# != 2 ]
then 
    echo "please check your command, parmeter number not valid!" # 参数数量不正确时提示
    echo "example:"
    echo "       ./go.sh --contracts_dir <tested_contracts_dir>" # 提供正确的命令示例
    exit -1 # 退出脚本
fi

# 检查第二个参数是否为有效目录
if [ ! -d $2  ]
then 
    echo "please check your command, '$2' not exists!" # 提示目录不存在
    echo "example:"
    echo "       ./go.sh --contracts_dir <tested_contracts_dir>" # 提供正确的命令示例
    exit -1 # 退出脚本
fi

# 获取测试合约目录的绝对路径
CONTRACT_DIR=$(cd $2 && pwd)

# 设置环境变量 CONTRACT_DIR
export CONTRACT_DIR
echo "Testing contracts from " $CONTRACT_DIR # 打印测试合约目录

# 启动 Geth 节点，并将日志输出到指定文件
nohup ./geth_run.sh >> $CONTRACT_DIR/fuzzer/reporter/geth_run.log 2>&1 &
sleep 60 # 等待 60 秒，确保 Geth 节点启动完成

# 返回工作目录
cd $DIR

# 启动测试器，并将日志输出到指定文件
nohup ./tester_run.sh >> $CONTRACT_DIR/fuzzer/reporter/tester_run.log 2>&1 &
sleep 300 # 等待 300 秒，确保测试器启动完成

# 返回工作目录
cd $DIR

# 启动模糊测试工具，并将日志输出到指定文件
./fuzzer_run.sh >> $CONTRACT_DIR/fuzzer/reporter/fuzzer_run.log 2>&1 

# 打印测试完成信息
echo "Test finished!"
echo "v_v..."
echo "Please go to $CONTRACT_DIR/fuzzer/reporter to see the results." # 提示用户查看测试结果