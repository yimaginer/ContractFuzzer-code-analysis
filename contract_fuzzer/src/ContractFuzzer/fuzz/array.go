package fuzz

import (
	"encoding/json" // 用于 JSON 编码和解码
	"fmt"           // 用于格式化输出
	"io"            // 用于输入输出操作
	"os"            // 用于文件操作
	"regexp"        // 用于正则表达式匹配
	"strconv"       // 用于字符串和数字之间的转换
	"strings"       // 用于字符串操作
)

// 定义正则表达式，用于匹配固定数组和动态数组的类型字符串
var fixReg = regexp.MustCompile("^(.*)\\[([\\d]+)\\]+$") // 匹配固定大小数组，如 "uint256[3]"
var DynReg = regexp.MustCompile("^(.*)\\[\\].*$")        // 匹配动态数组，如 "uint256[]"

// Fuzz 接口定义了模糊测试的核心方法
type Fuzz interface {
	fuzz(typestr string) string // 模糊测试方法，生成随机值
}

// FixedArray 表示固定大小的数组
type FixedArray struct {
	elem    Type          `json:"element_type"` // 数组元素的类型
	size    uint32        `json:"size"`         // 数组的大小
	str     string        `json:"description"`  // 描述字符串
	out     []interface{} `json:"fuzz_out"`     // 模糊测试生成的输出
	ostream io.Writer     `json:"-"`            // 输出流，用于写入文件
}

// newFixedArray 创建一个新的 FixedArray 对象
func newFixedArray(str string) *FixedArray {
	f := new(FixedArray)
	f.str = str
	match := fixReg.FindStringSubmatch(f.str) // 使用正则表达式匹配固定数组类型
	if len(match) != 0 {
		elemstr := match[1]               // 提取元素类型
		size, _ := strconv.Atoi(match[2]) // 提取数组大小
		elem, err := strToType(elemstr)   // 将元素类型字符串转换为 Type 对象
		if err != nil {
			fmt.Errorf("%s", err)
		}
		f.elem = elem
		f.size = uint32(size)
		f.out = make([]interface{}, 0)
	}
	return f
}

// String 方法将 FixedArray 转换为 JSON 字符串
func (f *FixedArray) String() string {
	buf, _ := json.Marshal(f)
	return string(buf)
}

// fuzz 方法生成固定大小数组的模糊测试值
// 每次生成一个数组项
func (f *FixedArray) fuzz() ([]interface{}, error) {
	var (
		isElem = true
		out    = make([]interface{}, 0)
		size   = f.size
	)
	for i := uint32(0); i < size; i++ {
		if m_out, err := f.elem.fuzz(isElem); err == nil {
			out = append(out, m_out[0])
		} else {
			return nil, err
		}
	}
	return []interface{}{out}, nil
}

// SetOstream 设置输出流，用于将模糊测试结果写入文件
func (f *FixedArray) SetOstream(file string) {
	if ostream, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666); err != nil {
		fmt.Printf("%s", FILE_OPEN_ERROR(err))
	} else {
		f.ostream = io.Writer(ostream)
	}
}

// Write 将数据写入输出流
func (f *FixedArray) Write(data []byte) {
	f.ostream.Write(data)
}

// DynamicArray 表示动态数组
type DynamicArray struct {
	elem Type          `json:"element_type"` // 数组元素的类型
	str  string        `json:"description"`  // 描述字符串
	out  []interface{} `json:"fuzz_out"`     // 模糊测试生成的输出
}

// newDynamicArray 创建一个新的 DynamicArray 对象
// "uint256[]" -> 空的动态数组
func newDynamicArray(str string) *DynamicArray {
	d := new(DynamicArray)
	d.str = str
	match := DynReg.FindStringSubmatch(d.str) // 使用正则表达式匹配动态数组类型
	if len(match) != 0 {
		elemstr := match[1]             // 提取元素类型
		elem, err := strToType(elemstr) // 将元素类型字符串转换为 Type 对象
		if err != nil {
			fmt.Errorf("%s", err)
		}
		d.elem = elem
		d.out = make([]interface{}, 0)
	}
	return d
}

// fuzz 方法生成动态数组的模糊测试值
func (d *DynamicArray) fuzz() ([]interface{}, error) {
	const ARRAY_SIZE_LIMIT = 10                                       // 动态数组的最大大小限制
	size := randintOne(1, ARRAY_SIZE_LIMIT)                           // 随机生成数组大小
	str_fixArray := fmt.Sprintf("%s[%d]", typeToString[d.elem], size) // 构造固定数组类型字符串
	fixArray := newFixedArray(str_fixArray)                           // 创建固定数组对象
	out, err := fixArray.fuzz()                                       // 调用固定数组的 fuzz 方法生成值
	return out, err
}

// String 方法将 DynamicArray 转换为 JSON 字符串
func (d *DynamicArray) String() string {
	buf, _ := json.Marshal(d)
	return string(buf)
}

// 定义类型常量，用于区分基本类型、固定数组和动态数组
const (
	Cfundemental  uint32 = iota // 基本类型
	CfixedArray                 // 固定数组
	CdynamicArray               // 动态数组
)

// getInfo 根据类型字符串获取类型信息
func getInfo(typestr string) (uint32, error) {
	typestr = strings.TrimSpace(typestr)

	if match := fixReg.MatchString(typestr); match == true {
		return CfixedArray, nil
	} else if match := DynReg.MatchString(typestr); match == true {
		return CdynamicArray, nil
	} else if v, err := strToType(typestr); err == nil {
		return Cfundemental, nil
	} else {
		return uint32(v), err
	}
}

// fuzz 函数根据类型字符串生成模糊测试值
func fuzz(str string) ([]interface{}, error) {
	v, err := getInfo(str)

	if err != nil {
		return nil, err
	} else {
		switch v {
		case Cfundemental:
			{
				f, _ := strToType(str)
				isElem := false
				out, _ := f.fuzz(isElem)
				return out, nil
			}
		case CfixedArray:
			{
				f := newFixedArray(str)
				out, _ := f.fuzz()
				return out, nil
			}
		case CdynamicArray:
			{
				d := newDynamicArray(str)
				out, _ := d.fuzz()
				return out, nil
			}
		default:
			return nil, ERR_UNKNOWN_COMPLEX_TYPE
		}
	}
}

// 该文件实现了对数组类型（包括固定数组和动态数组）的模糊测试功能，主要包括以下内容：
// 固定数组（FixedArray）:
// 		支持固定大小数组的解析和模糊测试。
// 		通过正则表达式匹配固定数组类型字符串（如 uint256[3]）。
// 动态数组（DynamicArray）:
// 		支持动态大小数组的解析和模糊测试。
// 		动态数组的大小在模糊测试时随机生成。
// 模糊测试核心逻辑:
// 		fuzz 函数根据类型字符串生成模糊测试值，支持基本类型、固定数组和动态数组。
// 输出功能:
// 		支持将模糊测试结果写入文件。
