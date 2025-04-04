// Copyright 2015 The go-ethereum Authors
// 本文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是开源的，你可以根据 GNU Lesser General Public License 的条款
// 自由地重新分发和修改它（版本 3 或更高版本）。
//
// go-ethereum 库是按“原样”分发的，没有任何明示或暗示的保证，
// 包括但不限于适销性或特定用途的适用性保证。
// 有关详细信息，请参阅 GNU Lesser General Public License。

package abi

import (
	"encoding/json" // 用于处理 JSON 数据
	"fmt"           // 格式化字符串
	"io"            // 输入输出操作
	"reflect"       // 反射操作
	"strings"       // 字符串操作

	"github.com/ethereum/go-ethereum/common" // go-ethereum 的通用工具包
)

// ABI 表示智能合约的上下文和可调用方法的信息。
// 它允许对函数调用进行类型检查，并根据 ABI 规范打包数据。
type ABI struct {
	Constructor Method            // 合约的构造函数
	Methods     map[string]Method // 合约中的方法集合
	Events      map[string]Event  // 合约中的事件集合
}

// JSON 从一个 JSON 数据流中解析出 ABI 接口。
// 如果解析失败，则返回错误。
func JSON(reader io.Reader) (ABI, error) {
	dec := json.NewDecoder(reader)

	var abi ABI
	if err := dec.Decode(&abi); err != nil {
		return ABI{}, err
	}

	return abi, nil
}

// Pack 根据 ABI 规范将方法名称和参数打包成字节数组。
// 方法调用的数据由方法 ID 和参数组成。
// 方法 ID 是方法签名哈希的前 4 个字节，参数是 32 字节对齐的数据。
func (abi ABI) Pack(name string, args ...interface{}) ([]byte, error) {
	var method Method

	// 如果方法名称为空，则表示是构造函数
	if name == "" {
		method = abi.Constructor
	} else {
		// 查找方法
		m, exist := abi.Methods[name]
		if !exist {
			return nil, fmt.Errorf("method '%s' not found", name)
		}
		method = m
	}

	// 打包参数
	arguments, err := method.pack(args...)
	if err != nil {
		return nil, err
	}

	// 如果是构造函数，直接返回参数
	if name == "" {
		return arguments, nil
	}

	// 否则返回方法 ID 和参数
	return append(method.Id(), arguments...), nil
}

// 这些变量用于在类型断言中确定特定类型。
var (
	r_interSlice = reflect.TypeOf([]interface{}{}) // 表示接口切片类型
	r_hash       = reflect.TypeOf(common.Hash{})   // 表示以太坊的哈希类型
	r_bytes      = reflect.TypeOf([]byte{})        // 表示字节切片类型
	r_byte       = reflect.TypeOf(byte(0))         // 表示单个字节类型
)

// 检查 output 是否为空。
// 确保目标变量是指针类型。
// 获取目标变量的值和类型。
// 根据方法输出参数的数量分两种情况处理：
// 		多个输出:
// 			如果目标是结构体，逐一匹配字段。
// 			如果目标是切片，逐一填充切片元素。
// 		单个输出:
// 			直接解包到目标变量中。
// 返回解包结果或错误。

// 根据 ABI 规范，将合约调用之后返回的原始字节数据解包到指定的目标变量中。它支持以下几种情况：
// Unpack 根据 ABI 规范将输出解包到指定的变量中。
// 参数 v 是解包的目标变量，name 是方法名称，output 是方法返回的数据。
func (abi ABI) Unpack(v interface{}, name string, output []byte) error {
	var method = abi.Methods[name]

	if len(output) == 0 {
		return fmt.Errorf("abi: unmarshalling empty output")
	}

	// 确保传入的变量是指针
	valueOf := reflect.ValueOf(v)
	if reflect.Ptr != valueOf.Kind() {
		return fmt.Errorf("abi: Unpack(non-pointer %T)", v)
	}

	var (
		value = valueOf.Elem() // 获取指针指向的值
		typ   = value.Type()   // 获取值的类型
	)

	// 如果方法有多个输出
	if len(method.Outputs) > 1 {
		switch value.Kind() {
		case reflect.Struct:
			// 将命名返回值与结构体字段匹配
			for i := 0; i < len(method.Outputs); i++ {
				marshalledValue, err := toGoType(i, method.Outputs[i], output)
				if err != nil {
					return err
				}
				reflectValue := reflect.ValueOf(marshalledValue)

				for j := 0; j < typ.NumField(); j++ {
					field := typ.Field(j)
					// TODO: 读取标签，例如 `abi:"fieldName"`
					if field.Name == strings.ToUpper(method.Outputs[i].Name[:1])+method.Outputs[i].Name[1:] {
						if err := set(value.Field(j), reflectValue, method.Outputs[i]); err != nil {
							return err
						}
					}
				}
			}
		case reflect.Slice:
			// 如果目标是切片，则解包到切片中
			if !value.Type().AssignableTo(r_interSlice) {
				return fmt.Errorf("abi: cannot marshal tuple in to slice %T (only []interface{} is supported)", v)
			}

			// 如果切片已经包含值，则设置这些值
			if value.Len() > 0 {
				if len(method.Outputs) > value.Len() {
					return fmt.Errorf("abi: cannot marshal in to slices of unequal size (require: %v, got: %v)", len(method.Outputs), value.Len())
				}

				for i := 0; i < len(method.Outputs); i++ {
					marshalledValue, err := toGoType(i, method.Outputs[i], output)
					if err != nil {
						return err
					}
					reflectValue := reflect.ValueOf(marshalledValue)
					if err := set(value.Index(i).Elem(), reflectValue, method.Outputs[i]); err != nil {
						return err
					}
				}
				return nil
			}

			// 创建一个新的切片并解包值
			z := reflect.MakeSlice(typ, 0, len(method.Outputs))
			for i := 0; i < len(method.Outputs); i++ {
				marshalledValue, err := toGoType(i, method.Outputs[i], output)
				if err != nil {
					return err
				}
				z = reflect.Append(z, reflect.ValueOf(marshalledValue))
			}
			value.Set(z)
		default:
			return fmt.Errorf("abi: cannot unmarshal tuple in to %v", typ)
		}

	} else {
		// 如果方法只有一个输出
		marshalledValue, err := toGoType(0, method.Outputs[0], output)
		if err != nil {
			return err
		}
		if err := set(value, reflect.ValueOf(marshalledValue), method.Outputs[0]); err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalJSON 从 JSON 数据中解析出 ABI 信息。
func (abi *ABI) UnmarshalJSON(data []byte) error {
	var fields []struct {
		Type      string
		Name      string
		Constant  bool
		Indexed   bool
		Anonymous bool
		Inputs    []Argument
		Outputs   []Argument
	}

	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	abi.Methods = make(map[string]Method)
	abi.Events = make(map[string]Event)
	for _, field := range fields {
		switch field.Type {
		case "constructor":
			abi.Constructor = Method{
				Inputs: field.Inputs,
			}
		// empty defaults to function according to the abi spec
		case "function", "":
			abi.Methods[field.Name] = Method{
				Name:    field.Name,
				Const:   field.Constant,
				Inputs:  field.Inputs,
				Outputs: field.Outputs,
			}
		case "event":
			abi.Events[field.Name] = Event{
				Name:      field.Name,
				Anonymous: field.Anonymous,
				Inputs:    field.Inputs,
			}
		}
	}

	return nil
}
