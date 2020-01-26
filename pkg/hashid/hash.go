package hashid

import "github.com/HFO4/cloudreve/pkg/conf"
import "github.com/speps/go-hashids"

// ID类型
const (
	ShareID = iota // 分享
	UserID         // 用户
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

// HashID 计算数据库内主键对应的HashID
func HashID(id uint, t int) string {
	v, _ := HashEncode([]int{int(id), t})
	return v
}
