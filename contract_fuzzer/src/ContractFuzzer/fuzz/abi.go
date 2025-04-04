package fuzz

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

// 计算笛卡尔积的第一个结果
func decarts_product_one(outs [][]interface{}) string {
	var vals = make([]interface{}, 0)
	if len(outs) == 0 {
		return ""
	}
	// 将第一个集合的元素添加到结果中
	for _, v := range outs[0] {
		vals = append(vals, stringify(v))
	}
	// 依次计算笛卡尔积
	for i := 1; i < len(outs); i++ {
		vals = product(vals, outs[i])
	}
	return vals[0].(string)
}

// 将值转换为字符串
func stringify(val interface{}) string {
	_, ok := val.(string)
	if ok {
		if strings.Contains(val.(string), "\"") {
			return val.(string)
		} else {
			return "\"" + val.(string) + "\""
		}
	} else {
		data, _ := json.Marshal(val)
		return string(data)
	}
}

// 计算两个集合的笛卡尔积
func product(A, B []interface{}) []interface{} {
	var rets = make([]interface{}, 0)
	for _, a := range A {
		for _, b := range B {
			rets = append(rets, stringify(a)+","+stringify(b))
		}
	}
	return rets
}

// 定义输入/输出元素的结构
type Elem struct {
	Name string        `json:"name,omitempty"` // 元素名称
	Type string        `json:"type"`           // 元素类型
	Out  []interface{} `json:"out,omitempty"`  // 元素的输出值
}

// 定义输入/输出的集合
type IOput []Elem

// 从 JSON 数据创建 IOput 对象
func newIOput(jsondata []byte) (*IOput, error) {
	var ioput = new(IOput)
	if err := json.Unmarshal(jsondata, ioput); err != nil {
		return nil, JSON_UNMARSHAL_ERROR(err)
	}
	return ioput, nil
}

// 将 IOput 转换为字符串
func (input *IOput) String() string {
	buf, _ := json.Marshal(input)
	return string(buf)
}

// 对 IOput 进行模糊测试
func (input *IOput) fuzz() (interface{}, error) {
	for i := range *input {
		elem := &(*input)[i]
		out, err := fuzz(elem.Type)
		if err != nil {
			return nil, err
		}
		elem.Out = out
	}
	var outs = make([][]interface{}, 0)
	for _, elem := range *input {
		outs = append(outs, elem.Out)
	}
	val := decarts_product_one(outs)
	return val, nil
}

// 定义函数的结构
type Function struct {
	Name     string `json:"name,omitempty"`     // 函数名称
	Type     string `json:"type"`               // 函数类型
	Inputs   IOput  `json:"inputs,omitempty"`   // 函数的输入
	Outputs  IOput  `json:"outputs,omitempty"`  // 函数的输出
	Payable  bool   `json:"payable"`            // 是否可支付
	Constant bool   `json:"constant,omitempty"` // 是否为常量
}

// 获取函数的签名
func (fun *Function) Sig() string {
	var elems = ([]Elem)(fun.Inputs)
	sig := fun.Name + "("
	for i, elem := range elems {
		if i == 0 {
			sig += elem.Type
		} else {
			sig += "," + elem.Type
		}
	}
	sig = sig + ")"
	return sig
}

// 获取函数的所有可能值
func (fun *Function) Values() []interface{} {
	var elems = ([]Elem)(fun.Inputs)
	var outs = make([][]interface{}, 0)

	for _, elem := range elems {
		outs = append(outs, elem.Out)
	}

	var vals = make([]interface{}, 0)
	if len(outs) == 0 {
		return nil
	} else {
		for _, v := range outs[0] {
			vals = append(vals, stringify(v))
		}
		for i := 1; i < len(outs); i++ {
			if i > 3 && len(outs[i]) > 2 {
				c := randintOne(len(outs[i]), 0)
				outs[i][0] = outs[i][c]
				c = randintOne(len(outs[i]), 0)
				outs[i][1] = outs[i][c]
				outs[i] = outs[i][:2]
			}
			vals = product(vals, outs[i])
		}
	}
	return vals
}

// 定义 ABI 的结构
type Abi []*Function

// 从 JSON 数据创建 ABI 对象
func newAbi(jsondata []byte) (*Abi, error) {
	var abi = new(Abi)
	if err := json.Unmarshal(jsondata, abi); err != nil {
		return nil, JSON_UNMARSHAL_ERROR(err)
	}
	return abi, nil
}

// type Function struct {
// 	Name     string `json:"name,omitempty"`     // 函数名称
// 	Type     string `json:"type"`               // 函数类型
// 	Inputs   IOput  `json:"inputs,omitempty"`   // 函数的输入
// 	Outputs  IOput  `json:"outputs,omitempty"`  // 函数的输出
// 	Payable  bool   `json:"payable"`            // 是否可支付
// 	Constant bool   `json:"constant,omitempty"` // 是否为常量
// }

// 输出 ABI 的值
func (abi *Abi) OutputValue(writer io.Writer) {
	funs := ([]*Function)(*abi)
	for _, fun := range funs {
		if fun != nil {
			sig := fun.Sig()
			values := fun.Values()
			log.Println(sig + ":" + string(len(values)))
			if len(values) != 0 {
				if len(values) > Global_fun_scale {
					out_values := make([]interface{}, 0, Global_fun_scale)
					for i := 0; i < Global_fun_scale; i++ {
						c := randintOne(len(values), 0)
						out_values = append(out_values, values[c])
					}
					values = out_values
				}
				for _, value := range values {
					writer.Write([]byte(sig + ":"))
					writer.Write([]byte("[" + value.(string) + "]"))
					writer.Write([]byte("\n"))
				}
			} else if len([]Elem(fun.Inputs)) == 0 {
				writer.Write([]byte(sig + "\n"))
			}
		}
	}
}

// 将 ABI 转换为字符串
func (abi *Abi) String() string {
	buf, _ := json.Marshal(abi)
	return string(buf)
}

// fuzz 函数的主要逻辑如下：

// 随机选择函数：
// 		从 ABI 中随机选择一个函数。
// 		如果函数类型不是 "function"，继续选择。
// 模糊测试输入参数：
// 		如果函数有输入参数，调用 fuzz 方法生成随机输入值。
// 		返回函数签名和生成的输入值。
// 无输入参数的处理：
// 		如果函数没有输入参数，直接返回函数签名。
// 错误处理：
// 		使用 recover 捕获运行时错误，防止程序崩溃。
// 		如果随机选择函数或模糊测试失败，返回 "0x0" 表示失败。

// 该函数对 ABI 中的函数进行模糊测试，随机选择一个函数并生成其输入值。如果函数没有输入参数，则直接返回函数签名。
func (abi *Abi) fuzz() (ret interface{}, ret_err error) {
	// 错误恢复机制
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			ret = nil
			ret_err = ERR_ABI_FUZZ_FAILED
		}
	}()
	// 将 Abi 对象中的函数列表转换为通用的 interface{} 切片，便于后续随机选择函数。
	funs := ([]*Function)(*abi)
	funcs := make([]interface{}, len(funs))
	for i := 0; i < len(funs); i++ {
		funcs[i] = funs[i]
	}
	var func_chose interface{}
	var err error
	var f *Function
	func_chose = nil

	// 	使用全局变量 g_func_Robin 的 Random_select 方法随机选择一个函数。
	// 	如果选择的函数类型不是 "function"，则继续随机选择，直到找到一个有效的函数或发生错误。
	for func_chose, err = g_func_Robin.Random_select(funcs); err == nil && func_chose.(*Function).Type != "function"; func_chose, err = g_func_Robin.Random_select(funcs) {
	}
	if err != nil {
		return "0x0", nil
	}
	// 将选择的函数转换为 *Function 类型，并检查其输入参数。
	// 如果函数有输入参数，则调用其 fuzz 方法生成随机输入值。
	f = func_chose.(*Function)
	if len(f.Inputs) > 0 {
		G_current_fun = f
		if ret, err := f.Inputs.fuzz(); err == nil {
			return fmt.Sprintf("%s:[%s]", f.Sig(), ret.(string)), nil
		} else {
			return "0x0", err
		}
	} else {
		return fmt.Sprintf("%s", f.Sig()), nil
	}
}
