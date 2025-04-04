// Copyright 2016 The go-ethereum Authors
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
	"math/big" // 用于处理大整数
	"reflect"  // 用于反射操作
	"strings"  // 用于字符串操作

	"github.com/ethereum/go-ethereum/common"      // go-ethereum 的通用工具包
	"github.com/ethereum/go-ethereum/common/math" // go-ethereum 的数学工具包
)

// packBytesSlice 将给定的字节切片打包为 [L, V] 的规范表示形式。
// L 表示字节切片的长度，V 表示字节切片的值。
func packBytesSlice(bytes []byte, l int) []byte {
	len := packNum(reflect.ValueOf(l))                               // 将长度打包为 32 字节
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...) // 将字节切片右填充为 32 字节对齐
}

// packElement 根据 ABI 规范打包给定的反射值。
// t 是类型信息，reflectValue 是需要打包的值。
func packElement(t Type, reflectValue reflect.Value) []byte {
	switch t.T {
	case IntTy, UintTy:
		// 打包整数类型
		return packNum(reflectValue)
	case StringTy:
		// 打包字符串类型
		return packBytesSlice([]byte(reflectValue.String()), reflectValue.Len())
	case AddressTy:
		// 打包地址类型
		return packAddress(reflectValue)
	case BoolTy:
		// 打包布尔类型
		return packBool(reflectValue)
	case BytesTy:
		// 打包字节类型
		bytevalues := packBytesTy(reflectValue)
		return packBytesSlice(bytevalues, reflectValue.Len())
	case FixedBytesTy, FunctionTy:
		// 打包固定字节或函数类型
		bytevalues := packBytesTy(reflectValue)
		return common.RightPadBytes(bytevalues, 32)
	}
	// 如果类型不支持，抛出错误
	panic("abi: fatal error")
}

// packNum 根据反射值打包数字类型。
// 支持无符号整数、有符号整数和十六进制字符串。
func packNum(value reflect.Value) []byte {
	switch value.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 打包无符号整数
		return U256(new(big.Int).SetUint64(value.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 打包有符号整数
		return U256(big.NewInt(value.Int()))
	default:
		// 处理十六进制字符串
		hexstr := value.Interface().(string)
		bint := big.Int{}
		bint.SetString(hexstr[2:], 16) // 将十六进制字符串转换为大整数
		return U256(&bint)
	}
}

// packAddress 打包以太坊地址类型。
// 地址必须是 40 个十六进制字符（不包括前缀 "0x"）。
func packAddress(value reflect.Value) []byte {
	hexstr := value.Interface().(string)
	if len(hexstr[2:]) != 40 {
		// 如果地址长度不正确，返回 nil
		return nil
	}
	// 左填充为 32 字节
	return common.LeftPadBytes(common.Hex2Bytes(hexstr[2:]), 32)
}

// packBool 打包布尔值。
// 支持布尔类型和字符串类型（"true"/"false"）。
func packBool(value reflect.Value) []byte {
	switch value.Kind() {
	case reflect.Bool:
		// 如果为 true，返回填充后的 1
		if value.Bool() {
			return math.PaddedBigBytes(common.Big1, 32)
		} else {
			// 如果为 false，返回填充后的 0
			return math.PaddedBigBytes(common.Big0, 32)
		}
	case reflect.String:
		// 处理字符串类型的布尔值
		str := value.Interface().(string)
		if strings.Compare(str, "true") == 0 {
			// 如果字符串为 "true"，返回填充后的 1
			return math.PaddedBigBytes(common.Big1, 32)
		} else {
			// 如果字符串为 "false"，返回填充后的 0
			return math.PaddedBigBytes(common.Big0, 32)
		}
	}
	// 如果类型不支持，返回 nil
	return nil
}

// packBytesTy 打包字节类型。
// 假设输入是一个十六进制字符串，去掉前缀 "0x" 后转换为字节数组。
func packBytesTy(value reflect.Value) []byte {
	hexstr := value.Interface().(string)
	// 去掉 "0x" 前缀并转换为字节数组
	return common.Hex2Bytes(hexstr[2:])
}
