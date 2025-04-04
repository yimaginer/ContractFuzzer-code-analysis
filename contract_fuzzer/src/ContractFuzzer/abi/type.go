// Copyright 2015 The go-ethereum Authors
// 本文件是 go-ethereum 库的一部分。
//
// go-ethereum 库是开源的，你可以根据 GNU Lesser General Public License 的条款
// 自由地重新分发和修改它（版本 3 或更高版本）。
//
// go-ethereum 库是按“原样”分发的，没有任何明示或暗示的保证，
// 包括但不限于适销性或特定用途的适用性保证。
// 有关详细信息，请参阅 GNU Lesser General Public License。

// 该文件通过正则表达式解析复杂的 ABI 类型，并提供了类型信息的封装和操作方法。
// 支持动态类型解析和值打包，适用于以太坊智能合约的 ABI 编码场景。
// 提供了灵活的接口和详细的错误处理机制，便于扩展和调试。

package abi

import (
	"fmt"     // 用于格式化字符串
	"reflect" // 用于反射操作
	"regexp"  // 用于正则表达式匹配
	"strconv" // 用于字符串和数字之间的转换
)

const (
	// 定义支持的 ABI 类型的常量
	IntTy        byte = iota // 有符号整数类型
	UintTy                   // 无符号整数类型
	BoolTy                   // 布尔类型
	StringTy                 // 字符串类型
	SliceTy                  // 切片类型
	AddressTy                // 地址类型
	FixedBytesTy             // 固定字节数组类型
	BytesTy                  // 动态字节数组类型
	HashTy                   // 哈希类型
	FixedPointTy             // 固定点数类型
	FunctionTy               // 函数类型
)

// Type 表示 ABI 支持的参数类型的反射信息
type Type struct {
	IsSlice, IsArray bool // 是否为切片或数组
	SliceSize        int  // 切片或数组的大小

	Elem *Type // 如果是复合类型（如数组或切片），表示其元素类型

	Kind reflect.Kind // Go 的反射类型
	Type reflect.Type // Go 的具体类型
	Size int          // 类型的大小（如 int256 的大小为 256）
	T    byte         // 自定义的类型标识符

	stringKind string // 保存未解析的字符串类型，用于生成签名
}

var (
	// fullTypeRegex 用于解析完整的 ABI 类型
	//
	// 类型格式可以是：
	// Input  = Type [ "[" [ Number ] "]" ] Name .
	// Type   = [ "u" ] "int" [ Number ] [ x ] [ Number ].
	//
	// 示例：
	// string     int       uint       fixed
	// string32   int8      uint8      uint[]
	// address    int256    uint256    fixed128x128[2]
	fullTypeRegex = regexp.MustCompile(`([a-zA-Z0-9]+)(\[([0-9]*)\])?`)

	// typeRegex 用于解析 ABI 子类型
	typeRegex = regexp.MustCompile("([a-zA-Z]+)(([0-9]+)(x([0-9]+))?)?")
)

// NewType 函数通过正则表达式解析输入字符串，递归构建 Type 对象。
// 对于复杂类型（如数组或切片），会递归解析其元素类型。
// 返回的 Type 对象包含完整的类型信息，包括是否为数组/切片、元素类型、大小等。
// 该函数适用于以太坊智能合约的 ABI 类型解析，支持多种复杂类型（如 uint256[]、fixed128x128[2] 等）。

// 输如: t := "uint256[]"
// 输出：
// Type: {IsSlice:true IsArray:false SliceSize:-1 Elem:0xc0000b8000 Kind:slice Type:<nil> Size:0 T:5 stringKind:uint256[]}
// Element Type: {IsSlice:false IsArray:false SliceSize:0 Elem:<nil> Kind:uint Type:uint Size:256 T:1 stringKind:uint256}

// 补充知识：
// 切片是 Go 语言的概念，用于表示动态大小的数组。
// 在以太坊中，动态数组是 ABI 类型的一部分，Go 语言使用切片来实现对动态数组的表示和操作。
// 切片的动态特性非常适合处理以太坊 ABI 中的动态数组，因此在以太坊相关的 Go 实现中，切片被广泛使用。

// 在 Go 语言中，动态数组通常用切片（slice）来表示。例如：
// uint256[] 在 Go 中可以表示为 []uint64。
// bytes 在 Go 中可以表示为 []byte。

// NewType 根据给定的字符串创建一个 ABI 类型的反射信息
func NewType(t string) (typ Type, err error) {
	// 使用正则表达式解析类型字符串
	res := fullTypeRegex.FindAllStringSubmatch(t, -1)[0]

	// 检查类型是否为切片或数组，并解析其类型
	switch {
	case res[3] != "":
		// 如果是数组，解析数组大小
		typ.SliceSize, _ = strconv.Atoi(res[3])
		typ.IsArray = true
	case res[2] != "":
		// 如果是切片，标记为切片类型
		typ.IsSlice, typ.SliceSize = true, -1
	case res[0] == "":
		// 如果类型解析失败，返回错误
		return Type{}, fmt.Errorf("abi: type parse error: %s", t)
	}

	// 如果是数组或切片，递归解析其元素类型
	if typ.IsArray || typ.IsSlice {
		sliceType, err := NewType(res[1])
		if err != nil {
			return Type{}, err
		}
		typ.Elem = &sliceType
		typ.stringKind = sliceType.stringKind + t[len(res[1]):]

		// 如果元素类型仍然是数组或切片，直接返回
		if typ.Elem.IsArray || typ.Elem.IsSlice {
			return typ, nil
		}
	}

	// 解析类型的具体信息（如类型名称和大小）
	parsedType := typeRegex.FindAllStringSubmatch(res[1], -1)[0]
	var varSize int
	if len(parsedType[3]) > 0 {
		varSize, err = strconv.Atoi(parsedType[2])
		if err != nil {
			return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
		}
	}

	// 如果是 int 或 uint 类型且未指定大小，默认为 256 位
	if varSize == 0 && (parsedType[1] == "int" || parsedType[1] == "uint") {
		varSize = 256
		t += "256"
	}

	// 如果不是数组或切片，设置类型的字符串表示
	if !(typ.IsArray || typ.IsSlice) {
		typ.stringKind = t
	}

	// 根据类型名称设置类型的具体信息
	switch parsedType[1] {
	case "int":
		typ.Kind, typ.Type = reflectIntKindAndType(false, varSize)
		typ.Size = varSize
		typ.T = IntTy
	case "uint":
		typ.Kind, typ.Type = reflectIntKindAndType(true, varSize)
		typ.Size = varSize
		typ.T = UintTy
	case "bool":
		typ.Kind = reflect.Bool
		typ.T = BoolTy
	case "address":
		typ.Kind = reflect.Array
		typ.Type = address_t
		typ.Size = 20
		typ.T = AddressTy
	case "string":
		typ.Kind = reflect.String
		typ.Size = -1
		typ.T = StringTy
	case "bytes":
		sliceType, _ := NewType("uint8")
		typ.Elem = &sliceType
		if varSize == 0 {
			typ.IsSlice = true
			typ.T = BytesTy
			typ.SliceSize = -1
		} else {
			typ.IsArray = true
			typ.T = FixedBytesTy
			typ.SliceSize = varSize
		}
	case "function":
		sliceType, _ := NewType("uint8")
		typ.Elem = &sliceType
		typ.IsArray = true
		typ.T = FunctionTy
		typ.SliceSize = 24
	default:
		// 如果类型不支持，返回错误
		return Type{}, fmt.Errorf("unsupported arg type: %s", t)
	}

	return
}

// String 实现 Stringer 接口，返回类型的字符串表示
func (t Type) String() (out string) {
	return t.stringKind
}

// pack 根据类型信息打包给定的值

// 假设输入: t:uint256[] ; 值 v 为 [1, 2, 3]
// 输出:
// 0000000000000000000000000000000000000000000000000000000000000003
// 0000000000000000000000000000000000000000000000000000000000000001
// 0000000000000000000000000000000000000000000000000000000000000002
// 0000000000000000000000000000000000000000000000000000000000000003
func (t Type) pack(v reflect.Value) ([]byte, error) {
	// 如果值是指针，先解引用
	v = indirect(v)

	// 如果是切片或数组，递归打包其元素
	if (t.IsSlice || t.IsArray) && t.T != BytesTy && t.T != FixedBytesTy && t.T != FunctionTy {
		var packed []byte
		for i := 0; i < v.Len(); i++ { // 切片或数组才有 len 之说
			val, err := t.Elem.pack(v.Index(i))
			if err != nil {
				return nil, err
			}
			packed = append(packed, val...)
		}
		if t.IsSlice {
			return packBytesSlice(packed, v.Len()), nil
		} else if t.IsArray {
			return packed, nil
		}
	}

	// 对单个值进行打包
	return packElement(t, v), nil
}

// requiresLengthPrefix 判断类型是否需要长度前缀
func (t Type) requiresLengthPrefix() bool {
	return t.T != FixedBytesTy && (t.T == StringTy || t.T == BytesTy || t.IsSlice)
}
