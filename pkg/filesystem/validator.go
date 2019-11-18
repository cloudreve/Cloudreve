package filesystem

import (
	"context"
	"github.com/HFO4/cloudreve/pkg/util"
	"path/filepath"
	"strings"
)

// 文件/路径名保留字符
var reservedCharacter = []string{"\\", "?", "*", "<", "\"", ":", ">", "/"}

// ValidateLegalName 验证文件名/文件夹名是否合法
func (fs *FileSystem) ValidateLegalName(ctx context.Context, name string) bool {
	for _, value := range reservedCharacter {
		if strings.Contains(name, value) {
			return false
		}
	}
	return true
}

// ValidateFileSize 验证上传的文件大小是否超出限制
func (fs *FileSystem) ValidateFileSize(ctx context.Context, size uint64) bool {
	return size <= fs.User.Policy.MaxSize
}

// ValidateCapacity 验证并扣除用户容量
func (fs *FileSystem) ValidateCapacity(ctx context.Context, size uint64) bool {
	if fs.User.DeductionCapacity(size) {
		return true
	}
	return false
}

// ValidateExtension 验证文件扩展名
func (fs *FileSystem) ValidateExtension(ctx context.Context, fileName string) bool {
	// 不需要验证
	if len(fs.User.Policy.OptionsSerialized.FileType) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	// 无扩展名时
	if len(ext) == 0 {
		return false
	}

	if util.ContainsString(fs.User.Policy.OptionsSerialized.FileType, ext[1:]) {
		return true
	}

	return false
}
