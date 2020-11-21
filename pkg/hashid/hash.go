package hashid

import (
	"errors"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/speps/go-hashids"
)

// ID类型
const (
	ShareID  = iota // 分享
	UserID          // 用户
	FileID          // 文件ID
	FolderID        // 目录ID
	TagID           // 标签ID
	PolicyID        // 存储策略ID
)

var (
	// ErrTypeNotMatch ID类型不匹配
	ErrTypeNotMatch = errors.New("ID类型不匹配")
)

// HashEncode 对给定数据计算HashID
func HashEncode(v []int) (string, error) {
	hd := hashids.NewData()
	hd.Salt = conf.SystemConfig.HashIDSalt

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}

	id, err := h.Encode(v)
	if err != nil {
		return "", err
	}
	return id, nil
}

// HashDecode 对给定数据计算原始数据
func HashDecode(raw string) ([]int, error) {
	hd := hashids.NewData()
	hd.Salt = conf.SystemConfig.HashIDSalt

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return []int{}, err
	}

	return h.DecodeWithError(raw)

}

// HashID 计算数据库内主键对应的HashID
func HashID(id uint, t int) string {
	v, _ := HashEncode([]int{int(id), t})
	return v
}

// DecodeHashID 计算HashID对应的数据库ID
func DecodeHashID(id string, t int) (uint, error) {
	v, _ := HashDecode(id)
	if len(v) != 2 || v[1] != t {
		return 0, ErrTypeNotMatch
	}
	return uint(v[0]), nil
}
