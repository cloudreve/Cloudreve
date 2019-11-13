package model

import (
	"cloudreve/pkg/util"
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
	// 表字段
	gorm.Model
	Email         string `gorm:"type:varchar(100);unique_index"`
	Nick          string `gorm:"size:50"`
	Password      string `json:"-"`
	Status        int
	GroupID       uint
	PrimaryGroup  int
	ActivationKey string `json:"-"`
	Storage       uint64
	LastNotify    *time.Time
	OpenID        string `json:"-"`
	TwoFactor     string `json:"-"`
	Delay         int
	Avatar        string
	Options       string `json:"-",gorm:"size:4096"`

	// 关联模型
	Group Group

	// 数据库忽略字段
	OptionsSerialized UserOption `gorm:"-"`
}

// UserOption 用户个性化配置字段
type UserOption struct {
	ProfileOn int    `json:"profile_on"`
	WebDAVKey string `json:"webdav_key"`
}

// GetUserByID 用ID获取用户
func GetUserByID(ID interface{}) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).First(&user, ID)
	return user, result.Error
}

// GetUserByEmail 用Email获取用户
func GetUserByEmail(email string) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).Where("email = ?", email).First(&user)
	return user, result.Error
}

// NewUser 返回一个新的空 User
func NewUser() User {
	options := UserOption{
		ProfileOn: 1,
	}
	optionsValue, _ := json.Marshal(&options)
	return User{
		Avatar:  "default",
		Options: string(optionsValue),
	}
}

// AfterFind 找到用户后的钩子
func (user *User) AfterFind() (err error) {
	// 解析用户设置到OptionsSerialized
	err = json.Unmarshal([]byte(user.Options), &user.OptionsSerialized)
	return err
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
