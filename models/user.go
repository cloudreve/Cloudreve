package model

import (
	"Cloudreve/pkg/serializer"
	"Cloudreve/pkg/util"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	// Active 账户正常状态
	Active = iota
	// NotActivicated 未激活
	NotActivicated
	// Baned 被封禁
	Baned
)

// User 用户模型
type User struct {
	gorm.Model
	Email         string `gorm:"type:varchar(100);unique_index"`
	Nick          string `gorm:"size:50"`
	Password      string
	Status        int
	Group         int
	PrimaryGroup  int
	ActivationKey string
	Storage       int64
	LastNotify    *time.Time
	OpenID        string
	TwoFactor     string
	Delay         int
	Avatar        string
	Options       string `gorm:"size:4096"`
}

// GetUser 用ID获取用户
func GetUser(ID interface{}) (User, error) {
	var user User
	result := DB.First(&user, ID)
	return user, result.Error
}

// NewUser 返回一个新的空 User
func NewUser() User {
	options := serializer.UserOption{
		ProfileOn: 1,
	}
	optionsValue, _ := json.Marshal(&options)
	return User{
		Avatar:  "default",
		Options: string(optionsValue),
	}
}

// CheckPassword 根据明文校验密码
func (user *User) CheckPassword(password string) (bool, error) {

	// 根据存储密码拆分为 Salt 和 Digest
	passwordStore := strings.Split(user.Password, ":")
	if len(passwordStore) != 2 {
		return false, errors.New("Unknown password type")
	}

	// todo 兼容V2/V1密码
	//计算 Salt 和密码组合的SHA1摘要
	hash := sha1.New()
	_, err := hash.Write([]byte(password + passwordStore[0]))
	bs := hex.EncodeToString(hash.Sum(nil))
	if err != nil {
		return false, err
	}

	return bs == passwordStore[1], nil
}

// SetPassword 根据给定明文设定 User 的 Password 字段
func (user *User) SetPassword(password string) error {
	//生成16位 Salt
	salt := util.RandStringRunes(16)

	//计算 Salt 和密码组合的SHA1摘要
	hash := sha1.New()
	_, err := hash.Write([]byte(password + salt))
	bs := hex.EncodeToString(hash.Sum(nil))

	if err != nil {
		return err
	}

	//存储 Salt 值和摘要， ":"分割
	user.Password = salt + ":" + string(bs)
	return nil
}
