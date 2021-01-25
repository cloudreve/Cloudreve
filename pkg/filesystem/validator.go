package filesystem

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

/* ==========
	 验证器
   ==========
*/

// 文件/路径名保留字符
var reservedCharacter = []string{"\\", "?", "*", "<", "\"", ":", ">", "/", "|"}

// ValidateLegalName 验证文件名/文件夹名是否合法
func (fs *FileSystem) ValidateLegalName(ctx context.Context, name string) bool {
	// 是否包含保留字符
	for _, value := range reservedCharacter {
		if strings.Contains(name, value) {
			return false
		}
	}

	// 是否超出长度限制
	if len(name) >= 256 {
		return false
	}

	// 是否为空限制
	if len(name) == 0 {
		return false
	}

	// 结尾不能是空格
	if strings.HasSuffix(name, " ") {
		return false
	}

	return true
}

// ValidateFileSize 验证上传的文件大小是否超出限制
func (fs *FileSystem) ValidateFileSize(ctx context.Context, size uint64) bool {
	if fs.User.Policy.MaxSize == 0 {
		return true
	}
	return size <= fs.User.Policy.MaxSize
}

// ValidateCapacity 验证并扣除用户容量
func (fs *FileSystem) ValidateCapacity(ctx context.Context, size uint64) bool {
	return fs.User.IncreaseStorage(size)
}

// ValidateExtension 验证文件扩展名
func (fs *FileSystem) ValidateExtension(ctx context.Context, fileName string) bool {
	// 不需要验证
	if len(fs.User.Policy.OptionsSerialized.FileType) == 0 {
		return true
	}

	return IsInExtensionList(fs.User.Policy.OptionsSerialized.FileType, fileName)
}

// IsInExtensionList 返回文件的扩展名是否在给定的列表范围内
func IsInExtensionList(extList []string, fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	// 无扩展名时
	if len(ext) == 0 {
		return false
	}

	if util.ContainsString(extList, ext[1:]) {
		return true
	}

	return false
}
