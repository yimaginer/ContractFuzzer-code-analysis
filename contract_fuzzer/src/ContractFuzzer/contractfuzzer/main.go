package main

import (
	// "ContractFuzzer/server"                   // 导入服务器模块
	// 之前的代码中使用了相对路径，使用相对路径的参照是本模块的go.mod文件所在位置,所以这个模块作者应该是单独开发完了之后放进来的
	"contract_fuzzer/src/ContractFuzzer/fuzz" // 导入模糊测试模块
	"contract_fuzzer/src/ContractFuzzer/server"
	"flag" // 用于解析命令行参数
	"log"  // 用于日志记录
)

var (
	// 定义命令行参数及其默认值
	abi_dir       = flag.String("abi_dir", "/verified_contract_abis", "input abi-dir")                                     // 指定存放智能合约 ABI 文件的目录
	out_dir       = flag.String("out_dir", "/verified_contract_abis_fuzz", "input out-dir")                                // 指定模糊测试结果的输出目录
	contract_list = flag.String("contract_list", "/list/config/contracts.list", "specify contract list for fuzzing input") // 指定要进行模糊测试的合约列表
	addr_seeds    = flag.String("addr_seeds", "/list/config/addressSeed.json", "specify address seedfile")                 // 指定地址种子文件
	int_seeds     = flag.String("int_seeds", "/list/config/intSeed.json", "specify int seedfile")                          // 指定整数种子文件
	uint_seeds    = flag.String("uint_seeds", "/list/config/uintSeed.json", "specify uint seedfile")                       // 指定无符号整数种子文件
	string_seeds  = flag.String("string_seeds", "/list/config/stringSeed.json", "specify string seedfile")                 // 指定字符串种子文件
	byte_seeds    = flag.String("byte_seeds", "/list/config/byteSeed.json", "specify bytes seedfile")                      // 指定字节种子文件
	bytes_seeds   = flag.String("bytes_seeds", "/list/config/bytesSeed.json", "specify bytes seedfile")                    // 指定字节数组种子文件
	fuzz_scale    = flag.Int("fuzz_scale", 5, "specify fuzz scale for each input param")                                   // 设置模糊测试的规模（深度或复杂度）
	input_scale   = flag.Int("input_scale", 8, "specify scale for fun")                                                    // 设置输入规模（生成测试输入的数量）
	fstart        = flag.Int("fstart", 2, "specify fuzz scale for each input param")                                       // 设置模糊测试的起始区块号
	fend          = flag.Int("fend", 2, "specify fuzz scale for each input param")                                         // 设置模糊测试的结束区块号
	addr_map      = flag.String("addr_map", "/list/config/addrmap.csv", "set addr_map")                                    // 指定地址映射文件
	abi_sigs_dir  = flag.String("abi_sigs_dir", "", "set abi_sigs_dir")                                                    // 指定 ABI 签名目录
	bin_sigs_dir  = flag.String("bin_sigs_dir", "", "set bin_sigs_dir")                                                    // 指定二进制签名目录
	listen_port   = flag.String("listen_port", "8888", "set listen_port")                                                  // 设置监听端口
	tester_port   = flag.String("tester_port", "http://localhost:6666", "set tester_port")                                 // 设置测试器的端口
	reporter      = flag.String("reporter", "/reporter", "specifiy results records direcotry")                             // 指定结果记录的目录
)

// 该文件是 ContractFuzzer 的主程序入口，
// 负责解析命令行参数、初始化模糊测试环境、启动模糊测试和服务器，并等待测试完成。
// 通过灵活的参数配置，用户可以指定测试的输入、规模和输出路径，
// 适用于对智能合约的安全性进行全面的模糊测试。

func main() {
	// 解析命令行参数
	flag.Parse()

	// 初始化模糊测试环境
	if err := fuzz.Init(*contract_list, *addr_seeds, *int_seeds, *uint_seeds, *string_seeds, *byte_seeds, *bytes_seeds, *fuzz_scale, *input_scale, *fstart, *fend, *addr_map, *abi_sigs_dir, *bin_sigs_dir, *listen_port, *tester_port); err != nil {
		log.Printf("%s\n", err) // 如果初始化失败，打印错误日志并退出
		return
	}

	// 启动模糊测试
	go fuzz.Start(*abi_dir, *out_dir)

	// 启动服务器，用于记录和管理模糊测试结果
	go server.Start(*addr_map, *reporter)

	// 等待模糊测试完成
	<-fuzz.G_finish
}
