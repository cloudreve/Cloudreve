package util

import "strings"

// DotPathToStandardPath 将","分割的路径转换为标准路径
func DotPathToStandardPath(path string) string {
	return "/" + strings.Replace(path, ",", "/", -1)
}

// FillSlash 给路径补全`/`
func FillSlash(path string) string {
	if path == "/" {
		return path
	}
	return path + "/"

}
