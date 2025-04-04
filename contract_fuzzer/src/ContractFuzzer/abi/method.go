// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"fmt"     // 用于格式化字符串
	"reflect" // 用于反射操作
	"strings" // 用于字符串操作

	"github.com/ethereum/go-ethereum/crypto" // 用于生成方法签名的哈希
	//"log" // 可选的日志包，用于调试
)

// Method 表示一个智能合约的方法。
// - `Name` 是方法的名称。
// - `Const` 表示方法是否为常量方法（即不需要发送交易即可调用）。
// - `Inputs` 是方法的输入参数列表。
// - `Outputs` 是方法的输出参数列表。
type Method struct {
	Name    string     // 方法名称
	Const   bool       // 是否为常量方法
	Inputs  []Argument // 输入参数列表
	Outputs []Argument // 输出参数列表
}

// pack 方法根据 ABI 规范将方法的输入参数打包为字节数组。
// - `args` 是输入参数的值。
// - 返回值是打包后的字节数组和可能的错误。
func (method Method) pack(args ...interface{}) ([]byte, error) {
	// 检查输入参数的数量是否匹配
	if len(args) != len(method.Inputs) {
		return nil, fmt.Errorf("argument count mismatch: %d for %d", len(args), len(method.Inputs))
	}

	// `variableInput` 用于存储需要长度前缀的变量类型（如字符串、字节数组等）的打包结果。
	var variableInput []byte

	// `ret` 用于存储最终的打包结果。
	var ret []byte

	// 遍历每个输入参数
	for i, a := range args {
		input := method.Inputs[i] // 获取对应的输入参数类型

		// 使用输入参数的类型信息打包参数值
		packed, err := input.Type.pack(reflect.ValueOf(a))
		if err != nil {
			return nil, fmt.Errorf("`%s` %v", method.Name, err)
		}

		// 检查参数是否需要长度前缀（如字符串、字节数组、切片等）
		if input.Type.requiresLengthPrefix() {
			// 计算偏移量（offset），即变量数据在最终打包结果中的位置
			offset := len(method.Inputs)*32 + len(variableInput)

			// 将偏移量打包并追加到结果中
			ret = append(ret, packNum(reflect.ValueOf(offset))...)

			// 将打包的变量数据追加到 `variableInput` 中
			variableInput = append(variableInput, packed...)
		} else {
			// 如果不需要长度前缀，直接将打包结果追加到 `ret` 中
			ret = append(ret, packed...)
		}
	}

	// 将所有变量数据追加到最终结果的末尾
	ret = append(ret, variableInput...)

	return ret, nil
}

// Sig 方法返回方法的字符串签名，符合 ABI 规范。
// - 示例：`function foo(uint32 a, int b)` 的签名为 `"foo(uint32,int256)"`。
// - 注意：`int` 会被替换为其规范表示 `int256`。
func (m Method) Sig() string {
	// 构造输入参数类型的字符串表示
	types := make([]string, len(m.Inputs))
	i := 0
	for _, input := range m.Inputs {
		types[i] = input.Type.String() // 获取参数类型的字符串表示
		i++
	}
	// 返回方法名称和参数类型的组合
	return fmt.Sprintf("%v(%v)", m.Name, strings.Join(types, ","))
}

// String 方法返回方法的完整字符串表示，包括名称、输入参数、输出参数和是否为常量方法。
// - 示例：`function foo(uint32 a, int b) constant returns (uint256)`
func (m Method) String() string {
	// 构造输入参数的字符串表示
	inputs := make([]string, len(m.Inputs))
	for i, input := range m.Inputs {
		inputs[i] = fmt.Sprintf("%v %v", input.Name, input.Type)
	}

	// 构造输出参数的字符串表示
	outputs := make([]string, len(m.Outputs))
	for i, output := range m.Outputs {
		if len(output.Name) > 0 {
			outputs[i] = fmt.Sprintf("%v ", output.Name)
		}
		outputs[i] += output.Type.String()
	}

	// 如果方法是常量方法，添加 `constant` 标志
	constant := ""
	if m.Const {
		constant = "constant "
	}

	// 返回完整的方法字符串表示
	return fmt.Sprintf("function %v(%v) %sreturns(%v)", m.Name, strings.Join(inputs, ", "), constant, strings.Join(outputs, ", "))
}

// Id 方法返回方法的 4 字节标识符（方法 ID）。
// - 方法 ID 是方法签名的 Keccak256 哈希的前 4 个字节。
// - 示例：`foo(uint32,int256)` 的方法 ID 为 `Keccak256("foo(uint32,int256)")[:4]`。
func (m Method) Id() []byte {
	return crypto.Keccak256([]byte(m.Sig()))[:4]
}

// 该文件实现了以太坊智能合约方法的核心功能，包括方法描述、参数打包、签名生成和方法 ID 计算。
// 提供了灵活的接口，支持静态和动态类型的参数处理，符合以太坊 ABI 规范。

// 该文件实现了以太坊智能合约中方法的描述和操作，包括：

// 方法的定义:
// 使用 Method 结构体表示智能合约的方法，包括名称、是否为常量、输入参数和输出参数。
// 方法的打包:
// pack 方法根据 ABI 规范将输入参数打包为字节数组，支持静态和动态类型。
// 方法签名:
// Sig 方法生成方法的字符串签名，符合 ABI 规范。
// 方法字符串表示:
// String 方法返回方法的完整字符串表示，包括名称、参数和返回值。
// 方法 ID:
// Id 方法生成方法的 4 字节标识符，用于以太坊交易调用。
