package fuzz

import (
	"os"
	"bufio"
	"strings"
	"log"
	"io"
	"fmt"
	"net/http"
	"net/url"
	abi_gen "ContractFuzzer/abi"
)
var (
	error_log = "/list/error-line.log"
)
var transport = http.Transport{
	DisableKeepAlives: false,
	}
var Client = http.Client{Transport:&transport}
var (
	Global_contractList []string=[]string{""}
	Global_addrSeed  string = ""
	Global_intSeed string = ""
	Global_uintSeed string = ""
	Global_stringSeed string = ""
	Global_byteSeed string = ""
	Global_bytesSeed string = ""
	Global_scale  int = 2
	Global_fun_scale int = 8
	Global_fstart int = 0
	Global_fend int  = 0
	Global_addr_map = ""
	Global_abi_sigs_dir = ""
	Global_bin_sigs_dir = ""
	Global_listen_port = ""
	Global_tester_port = ""
	GlobalADDR_MAP = make(map[string]string)
	GlobalFUNSIG_CONTRACT_MAP = make(map[string][]string)
)

func createADDR_MAP() error {
    // 打开地址映射文件
    f, err := os.Open(Global_addr_map)
    defer func() {
        f.Close() // 确保文件在函数结束时关闭
    }()
    if err != nil {
        return err // 如果文件打开失败，返回错误
    }

    // 创建一个缓冲读取器
    buf := bufio.NewReader(f)
    for {
        // 按行读取文件内容
        line, err := buf.ReadString('\n')
        if err != nil {
            // 如果读取失败，将错误写入日志文件
            f, _ := os.OpenFile(error_log, os.O_CREATE|os.O_APPEND, 0666)
            f.Write([]byte(err.Error()))
            f.Close()
        }

        if err != nil {
            if err == io.EOF {
                return nil // 如果读取到文件末尾，正常退出
            }
            return err // 如果发生其他错误，返回错误
        }

        if line == "" {
            return nil // 如果行为空，跳过
        }

        // 去除行首尾的空格
        line = strings.TrimSpace(line)

        // 按逗号分割行内容，提取地址和名称
        str2 := strings.Split(line, ",")
        addr := strings.TrimSpace(str2[0]) // 地址
        name := strings.TrimSpace(str2[1]) // 名称

        // 将地址和名称存储到全局地址映射表中
        GlobalADDR_MAP[name] = addr
    }
    return nil
}

func getFUNSIG_CONTRACT_by_file(file string) error {
    // 打开指定的 ABI 签名文件
    f, err := os.Open(Global_abi_sigs_dir + "/" + file)
    defer func() {
        f.Close() // 确保文件在函数结束时关闭
    }()
    if err != nil {
        return err // 如果文件打开失败，返回错误
    }

    // 提取文件名中的合约名称
    strs := strings.Split(file, "/")
    name := strings.Split(strs[len(strs)-1], ".")[0]

    // 创建一个缓冲读取器
    buf := bufio.NewReader(f)
    for {
        // 按行读取文件内容
        line, err := buf.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                return nil // 如果读取到文件末尾，正常退出
            }
            return err // 如果发生其他错误，返回错误
        }

        if line == "" {
            return nil // 如果行为空，跳过
        }

        // 去除行首尾的空格
        line = strings.TrimSpace(line)

        // 按冒号分割行内容，提取函数签名
        str2 := strings.Split(line, ":")
        fun_sig := str2[0] // 函数签名

        // 将函数签名与合约名称存储到全局映射表中
        _, found := GlobalFUNSIG_CONTRACT_MAP[fun_sig]
        if !found {
            GlobalFUNSIG_CONTRACT_MAP[fun_sig] = []string{name}
        } else {
            GlobalFUNSIG_CONTRACT_MAP[fun_sig] = append(GlobalFUNSIG_CONTRACT_MAP[fun_sig], name)
        }
    }
    return nil
}

func createFUNSIG_CONTRACT_MAP() error {
    // 读取 ABI 签名目录中的所有文件
    files, err := readDir(Global_abi_sigs_dir)
    if err != nil {
        return err // 如果读取目录失败，返回错误
    }

    // 遍历每个文件，生成函数签名与合约的映射
    for _, file := range files {
        if err := getFUNSIG_CONTRACT_by_file(file); err != nil {
            return fmt.Errorf("createFUNSIG_CONTRACT_MAP.getFUNSIG_CONTRACT_by_file %s: %s", file, err)
        }
    }
    return nil
}

func setG_current_bin_fun_sigs() error {
    // 打开当前合约的二进制签名文件
    f, err := os.Open(Global_bin_sigs_dir + "/" + G_current_contract.(string) + ".bin.sig")
    defer func() {
        f.Close() // 确保文件在函数结束时关闭
    }()
    if err != nil {
        return err // 如果文件打开失败，返回错误
    }

    // 创建一个缓冲读取器
    buf := bufio.NewReader(f)
    for {
        // 按行读取文件内容
        line, err := buf.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                return nil // 如果读取到文件末尾，正常退出
            }
            return err // 如果发生其他错误，返回错误
        }

        if line == "" {
            return nil // 如果行为空，跳过
        }

        // 去除行首尾的空格
        line = strings.TrimSpace(line)

        // 按冒号分割行内容，提取函数签名和内部调用签名
        str2 := strings.Split(line, ":")
        fun_sig := str2[0] // 函数签名
        innercall_funsigs := strings.Split(str2[1], " ") // 内部调用签名列表

        // 遍历内部调用签名列表，将其添加到全局变量 G_current_bin_fun_sigs 中
        for _, sig := range innercall_funsigs {
            _, found := G_current_bin_fun_sigs[fun_sig]
            if !found {
                G_current_bin_fun_sigs[fun_sig] = []string{sig}
            } else {
                G_current_bin_fun_sigs[fun_sig] = append(G_current_bin_fun_sigs[fun_sig], sig)
            }
        }
    }
    return nil
}

func setG_current_abi_sigs() error {
    // 打开当前合约的 ABI 签名文件
    f, err := os.Open(Global_abi_sigs_dir + "/" + G_current_contract.(string) + ".abi")
    if err != nil {
        return err // 如果文件打开失败，返回错误
    }

    // 创建一个缓冲读取器
    buf := bufio.NewReader(f)
    for {
        // 按行读取文件内容
        line, err := buf.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                return nil // 如果读取到文件末尾，正常退出
            }
            return err // 如果发生其他错误，返回错误
        }

        if line == "" {
            return nil // 如果行为空，跳过
        }

        // 去除行首尾的空格
        line = strings.TrimSpace(line)

        // 按冒号分割行内容，提取函数签名和函数签名字符串
        str2 := strings.Split(line, ":")
        fun_sig := str2[0] // 函数签名
        fun_sig_str := strings.TrimSpace(str2[1]) // 函数签名字符串

        // 将函数签名字符串与函数签名存储到全局变量 G_current_abi_sigs 中
        G_current_abi_sigs[fun_sig_str] = fun_sig
    }
    return nil
}
func  Init(contractListPath ,addrSeed,intSeed,uintSeed,stringSeed,byteSeed,
	 bytesSeed string,scale ,fun_scale, fstart,fend int,addr_map, abi_sigs_dir,bin_sigs_dir string,
	 listen_port,tester_port string)(error){
	Global_contractList = make([]string,0,0)
	if contractListPath != "null"{
		file,err:=os.Open(contractListPath)
		if err!=nil{
			return err
		}
		reader := bufio.NewReader(file)
		for contract,e:=reader.ReadString('\n');e==nil;contract,e=reader.ReadString('\n'){
			Global_contractList = append(Global_contractList,strings.Trim(contract,"\n")+".abi")
		}
	}
	Global_addrSeed = addrSeed
	Global_bytesSeed = bytesSeed
	Global_intSeed = intSeed
	Global_stringSeed = stringSeed
	Global_uintSeed = uintSeed
	Global_scale = scale
	Global_fstart = fstart
	Global_fend  = fend
	Global_fun_scale = fun_scale
	Global_addr_map = addr_map
	Global_abi_sigs_dir = abi_sigs_dir
	Global_bin_sigs_dir = bin_sigs_dir
	Global_listen_port = listen_port
	Global_tester_port = tester_port
	createADDR_MAP()
	createFUNSIG_CONTRACT_MAP()
	return nil
}
func sendMsg2RunnerMonitor(address string, msgs []string)(bool){
	values := url.Values{"address":[]string{address},"msg":msgs}
	go func(){
			// res,_ := Client.Get("http://localhost:6666/runnerMonitor?"+values.Encode())
			res,_ := Client.Get(Global_tester_port+"/runnerMonitor?"+values.Encode())
			// log.Println(res)
			defer func(){
				if err:= recover();err!=nil{
					log.Println(err)
				}else{
					res.Body.Close()
				}
			}()
	}()
	return true
}
var(
	rand_case_ranges = []interface{}{20,25,30,35,40}
    rand_case_scales = []interface{}{6,7,8,9,10}
)
var(
	RAND_CASE_RANGE = 10
	RAND_CASE_SCALE = 10
)
// no buffered channel.
// synchronized.
var(
	G_stop = make(chan bool,0)
	G_start = make(chan bool,0)
	G_finish = make(chan bool,0)
	G_sig_continue = make(chan bool,0)
) 

func Start(dir string, outdir string) error {
    defer func() {
        // 捕获异常，防止程序崩溃
        if err := recover(); err != nil {
            log.Println(err)
            printCallStackIfError() // 打印调用栈信息（如果有）
        }
    }()

    // 创建输出目录，如果目录已存在则忽略错误
    if err := os.Mkdir(outdir, 0777); err != nil {
        if !os.IsExist(err) {
            return err
        }
    }

    // 读取输入目录中的所有文件
    files, err := readDir(dir)
    if err != nil {
        return err
    }

    // 如果全局合约列表为空，则使用输入目录中的文件
    if len(Global_contractList) == 0 {
        // 遍历指定的区块范围内的文件
        for i := Global_fstart; i < Global_fend && i < len(files)-1; i++ {
            file := files[i]
            path := dir + "/" + file

            // 设置当前合约名称和地址
            G_current_contract = strings.TrimSpace(strings.Split(file, ".")[0])
            current_contract_address := GlobalADDR_MAP[G_current_contract.(string)]

            // 设置当前合约的二进制和 ABI 函数签名
            setG_current_bin_fun_sigs()
            setG_current_abi_sigs()

            // 读取合约文件内容
            data, err := readFile(path)
            if err != nil {
                continue
            }

            // 初始化 ABI 对象
            abi, err := newAbi(data)
            if err != nil {
                continue
            }

            // 随机选择测试范围（RAND_CASE_RANGE）
            if bid, err := g_robin.Random_select(rand_case_ranges); err == nil {
                RAND_CASE_RANGE = bid.(int)
            } else {
                continue
            }

            // 遍历测试范围内的每个测试用例
            for no := 0; no < RAND_CASE_RANGE; no++ {
                funs := make([]string, 0) // 存储函数签名
                msgs := make([]string, 0) // 存储生成的消息

                // 随机选择测试规模（RAND_CASE_SCALE）
                if bid, err := g_robin.Random_select(rand_case_scales); err == nil {
                    RAND_CASE_SCALE = bid.(int)
                } else {
                    continue
                }

                invalid_no := 0 // 无效测试用例计数

                // 遍历测试规模内的每个输入
                for i := 0; i < RAND_CASE_SCALE; i++ {
                    // 调用 ABI 的模糊测试方法生成输入
                    if ret, err := abi.fuzz(); err == nil {
                        log.Println(ret)
                        if strings.Contains(ret.(string), "0x0") == true {
                            // 如果生成的消息包含 "0x0"，则添加占位符消息
                            msgs = append(msgs, "0xcaffee")
                        } else {
                            // 解析生成的消息
                            if hex_str, err := abi_gen.Parse_GenMsg(ret.(string)); err == nil {
                                log.Println(hex_str)
                                msgs = append(msgs, hex_str)
                                // 截取函数签名（前 10 个字符）
                                if len(hex_str) > 10 {
                                    funs = append(funs, hex_str[:10])
                                } else {
                                    funs = append(funs, hex_str)
                                }
                            }
                        }
                    } else {
                        // 如果生成失败，计数无效测试用例
                        invalid_no++
                        continue
                    }
                }

                // 更新有效测试用例数量
                RAND_CASE_SCALE -= invalid_no
                if RAND_CASE_RANGE == 0 {
                    continue
                }

                // 通知开始测试
                G_start <- true
                <-G_sig_continue // 等待信号继续

                // 将生成的消息发送到 Runner Monitor
                sendMsg2RunnerMonitor(current_contract_address, msgs)

                // 清空消息和函数签名列表
                msgs = make([]string, 0)
                funs = make([]string, 0)

                // 如果到达测试范围的最后一个用例，退出循环
                if no == RAND_CASE_RANGE-1 {
                    break
                }

                // 如果收到停止信号，退出循环
                if c := <-G_stop; c == true {
                    break
                }
            }
        }

        // 通知测试完成
        G_finish <- true
    } else {
        // 如果全局合约列表不为空，则使用列表中的文件
        for _, file := range Global_contractList {
            path := dir + "/" + file

            // 读取合约文件内容
            data, err := readFile(path)
            if err != nil {
                continue
            }

            // 设置当前合约名称和地址
            G_current_contract = strings.TrimSpace(strings.Split(file, ".")[0])
            current_contract_address := GlobalADDR_MAP[G_current_contract.(string)]

            // 设置当前合约的二进制和 ABI 函数签名
            setG_current_bin_fun_sigs()
            setG_current_abi_sigs()

            // 初始化 ABI 对象
            abi, err := newAbi(data)
            if err != nil {
                continue
            }

            // 随机选择测试范围（RAND_CASE_RANGE）
            if bid, err := g_robin.Random_select(rand_case_ranges); err == nil {
                RAND_CASE_RANGE = bid.(int)
            } else {
                continue
            }

            // 遍历测试范围内的每个测试用例
            for no := 0; no < RAND_CASE_RANGE; no++ {
                funs := make([]string, 0) // 存储函数签名
                msgs := make([]string, 0) // 存储生成的消息

                // 随机选择测试规模（RAND_CASE_SCALE）
                if bid, err := g_robin.Random_select(rand_case_scales); err == nil {
                    RAND_CASE_SCALE = bid.(int)
                } else {
                    continue
                }

                invalid_no := 0 // 无效测试用例计数

                // 遍历测试规模内的每个输入
                for i := 0; i < RAND_CASE_SCALE; i++ {
                    // 调用 ABI 的模糊测试方法生成输入
                    if ret, err := abi.fuzz(); err == nil {
                        log.Println(ret)
                        if strings.Contains(ret.(string), "0x0") == true {
                            // 如果生成的消息包含 "0x0"，则添加占位符消息
                            msgs = append(msgs, "0xC0FFEE")
                        } else {
                            // 解析生成的消息
                            if hex_str, err := abi_gen.Parse_GenMsg(ret.(string)); err == nil {
                                log.Println(hex_str)
                                msgs = append(msgs, hex_str)
                                // 截取函数签名（前 10 个字符）
                                if len(hex_str) > 10 {
                                    funs = append(funs, hex_str[:10])
                                } else {
                                    funs = append(funs, hex_str)
                                }
                            }
                        }
                    } else {
                        // 如果生成失败，计数无效测试用例
                        invalid_no++
                        continue
                    }
                }

                // 更新有效测试用例数量
                RAND_CASE_SCALE -= invalid_no
                if RAND_CASE_RANGE == 0 {
                    continue
                }

                // 通知开始测试
                G_start <- true
                <-G_sig_continue // 等待信号继续

                // 将生成的消息发送到 Runner Monitor
                sendMsg2RunnerMonitor(current_contract_address, msgs)

                // 清空消息和函数签名列表
                msgs = make([]string, 0)
                funs = make([]string, 0)

                // 如果到达测试范围的最后一个用例，退出循环
                if no == RAND_CASE_RANGE-1 {
                    break
                }

                // 如果收到停止信号，退出循环
                if c := <-G_stop; c == true {
                    break
                }
            }
        }

        // 通知测试完成
        G_finish <- true
    }

    return nil
}