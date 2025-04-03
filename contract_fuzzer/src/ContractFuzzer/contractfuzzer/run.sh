#!/bin/sh
# 启动 ContractFuzzer 的模糊测试脚本

# 执行 Fuzzer_fuzz 程序，并传递相关参数
./Fuzzer_fuzz \
    -abi_dir /home/liuye/tested_contracts/abis \ # 指定存放智能合约 ABI 文件的目录
    -out_dir /home/liuye/tested_contracts/autofuzz \ # 指定模糊测试结果的输出目录
    -abi_sigs_dir /home/liuye/tested_contracts/abi_sig \ # 指定存放 ABI 签名的目录
    -bin_sigs_dir /home/liuye/tested_contracts/sig \ # 指定存放二进制签名的目录
    -fuzz_scale 5 \ # 设置模糊测试的规模（模糊测试的深度或复杂度）
    -input_scale 10 \ # 设置输入规模（生成测试输入的数量）
    -fstart 6000 \ # 设置模糊测试的起始区块号
    -fend 7000 \ # 设置模糊测试的结束区块号
    -addr_map /home/liuye/resource/addrmap.csv # 指定地址映射文件，用于解析合约地址