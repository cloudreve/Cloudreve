package util

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	RandomVariantAll = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	RandomLowerCases = []rune("1234567890abcdefghijklmnopqrstuvwxyz")
)

// RandStringRunes 返回随机字符串
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = RandomVariantAll[rand.Intn(len(RandomVariantAll))]
	}
	return string(b)
}

// RandString returns random string in given length and variant
func RandString(n int, variant []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = variant[rand.Intn(len(variant))]
	}
	return string(b)
}

// ContainsUint 返回list中是否包含
func ContainsUint(s []uint, e uint) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// IsInExtensionList 返回文件的扩展名是否在给定的列表范围内
func IsInExtensionList(extList []string, fileName string) bool {
	ext := Ext(fileName)
	// 无扩展名时
	if len(ext) == 0 {
		return false
	}

	if ContainsString(extList, ext) {
		return true
	}

	return false
}

// IsInExtensionList 返回文件的扩展名是否在给定的列表范围内
func IsInExtensionListExt(extList []string, ext string) bool {
	// 无扩展名时
	if len(ext) == 0 {
		return false
	}

	if ContainsString(extList, ext) {
		return true
	}

	return false
}

// ContainsString 返回list中是否包含
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Replace 根据替换表执行批量替换
func Replace(table map[string]string, s string) string {
	for key, value := range table {
		s = strings.Replace(s, key, value, -1)
	}
	return s
}

// BuildRegexp 构建用于SQL查询用的多条件正则
func BuildRegexp(search []string, prefix, suffix, condition string) string {
	var res string
	for key, value := range search {
		res += prefix + regexp.QuoteMeta(value) + suffix
		if key < len(search)-1 {
			res += condition
		}
	}
	return res
}

// BuildConcat 根据数据库类型构建字符串连接表达式
func BuildConcat(str1, str2 string, DBType string) string {
	switch DBType {
	case "mysql":
		return "CONCAT(" + str1 + "," + str2 + ")"
	default:
		return str1 + "||" + str2
	}
}

// SliceIntersect 求两个切片交集
func SliceIntersect(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 1 {
			nn = append(nn, v)
		}
	}
	return nn
}

// SliceDifference 求两个切片差集
func SliceDifference(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	inter := SliceIntersect(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}

	for _, value := range slice1 {
		times, _ := m[value]
		if times == 0 {
			nn = append(nn, value)
		}
	}
	return nn
}

// WithValue inject key-value pair into request context.
func WithValue(c *gin.Context, key any, value any) {
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), key, value))
}

// BoolToString transform bool to string
func BoolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func EncodeTimeFlowString(str string, timeNow int64) string {
	timeNow = timeNow / 1000
	timeDigits := []int{}
	timeDigitIndex := 0

	if len(str) == 0 {
		return ""
	}

	str = fmt.Sprintf("%d|%s", timeNow, str)

	res := make([]int32, 0, utf8.RuneCountInString(str))
	for timeNow > 0 {
		timeDigits = append(timeDigits, int(timeNow%int64(10)))
		timeNow = timeNow / 10
	}

	add := false
	for pos, rune := range str {
		// take single digit with index timeDigitIndex from timeNow
		newIndex := pos
		if add {
			newIndex = pos + timeDigits[timeDigitIndex]*timeDigitIndex
		} else {
			newIndex = 2*timeDigitIndex*timeDigits[timeDigitIndex] - pos
		}

		if newIndex < 0 {
			newIndex = newIndex * -1
		}

		res = append(res, rune)
		newIndex = newIndex % len(res)

		res[newIndex], res[len(res)-1] = res[len(res)-1], res[newIndex]

		add = !add
		// Add timeDigitIndex by 1, but does not exceed total digits in timeNow
		timeDigitIndex++
		timeDigitIndex = timeDigitIndex % len(timeDigits)
	}

	return string(res)
}

func DecodeTimeFlowStringTime(str string, timeNow int64) string {
	timeNow = timeNow / 1000
	timeDigits := []int{}

	if len(str) == 0 {
		return ""
	}

	for timeNow > 0 {
		timeDigits = append(timeDigits, int(timeNow%int64(10)))
		timeNow = timeNow / 10
	}

	res := make([]int32, utf8.RuneCountInString(str))
	secret := []rune(str)
	add := false
	if len(secret)%2 == 0 {
		add = true
	}
	timeDigitIndex := (len(secret) - 1) % len(timeDigits)
	for pos := range secret {
		// take single digit with index timeDigitIndex from timeNow
		newIndex := len(res) - 1 - pos
		if add {
			newIndex = newIndex + timeDigits[timeDigitIndex]*timeDigitIndex
		} else {
			newIndex = 2*timeDigitIndex*timeDigits[timeDigitIndex] - newIndex
		}

		if newIndex < 0 {
			newIndex = newIndex * -1
		}

		newIndex = newIndex % len(secret)

		res[len(res)-1-pos] = secret[newIndex]
		secret[newIndex], secret[len(res)-1-pos] = secret[len(res)-1-pos], secret[newIndex]
		secret = secret[:len(secret)-1]

		add = !add
		// Add timeDigitIndex by 1, but does not exceed total digits in timeNow
		timeDigitIndex--
		if timeDigitIndex < 0 {
			timeDigitIndex = len(timeDigits) - 1
		}
	}

	return string(res)
}

func ToPtr[T any](v T) *T {
	return &v
}
