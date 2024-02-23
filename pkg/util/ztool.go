package util

import "strings"

type (
	// 可取长度类型
	LenAble interface{ string | []any | chan any }
)

// 计算切片元素总长度
/*
 传入字符串切片, 返回其所有元素长度之和
 e.g. LenArray({`ele3`,`ele2`,`ele1`}) => 12
*/
func LenArray[T LenAble](a []T) int {
	var o int
	for i, r := 0, len(a); i < r; i++ {
		o += len(a[i])
	}
	return o
}

// 字符串快速拼接
/*
 传入多个字符串参数, 返回拼接后的结果
 e.g: StrConcat("str1", "str2", "str3") => "str1str2str3"
*/
func StrConcat(a ...string) string {
	var b strings.Builder
	b.Grow(LenArray(a))
	for i, r := 0, len(a); i < r; i++ {
		b.WriteString(a[i])
	}
	return b.String()
}
