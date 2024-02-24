package model

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	// Active 账户正常状态
	Active = iota
	// NotActivicated 未激活
	NotActivicated
	// Baned 被封禁
	Baned
	// OveruseBaned 超额使用被封禁
	OveruseBaned
)

// User 用户模型
type User struct {
	// 表字段
	gorm.Model
	Email           string `gorm:"type:varchar(100);unique_index"`
	Nick            string `gorm:"size:50"`
	Password        string `json:"-"`
	Status          int
	GroupID         uint
	Storage         uint64
	OpenID          string
	TwoFactor       string
	Avatar          string
	Options         string `json:"-" gorm:"size:4294967295"`
	Authn           string `gorm:"size:4294967295"`
	Score           int
	PreviousGroupID uint       // 初始用户组
	GroupExpires    *time.Time // 用户组过期日期
	NotifyDate      *time.Time // 通知超出配额时的日期
	Phone           string

	// 关联模型
	Group Group `gorm:"save_associations:false:false"`

	// 数据库忽略字段
	OptionsSerialized UserOption `gorm:"-"`
}

func init() {
	gob.Register(User{})
}

// UserOption 用户个性化配置字段
type UserOption struct {
	ProfileOff      bool   `json:"profile_off,omitempty"`
	PreferredPolicy uint   `json:"preferred_policy,omitempty"`
	PreferredTheme  string `json:"preferred_theme,omitempty"`
}

// Root 获取用户的根目录
func (user *User) Root() (*Folder, error) {
	var folder Folder
	err := DB.Where("parent_id is NULL AND owner_id = ?", user.ID).First(&folder).Error
	return &folder, err
}

// DeductionStorage 减少用户已用容量
func (user *User) DeductionStorage(size uint64) bool {
	if size == 0 {
		return true
	}
	if size <= user.Storage {
		user.Storage -= size
		DB.Model(user).Update("storage", gorm.Expr("storage - ?", size))
		return true
	}
	// 如果要减少的容量超出已用容量，则设为零
	user.Storage = 0
	DB.Model(user).Update("storage", 0)

	return false
}

// IncreaseStorage 检查并增加用户已用容量
func (user *User) IncreaseStorage(size uint64) bool {
	if size == 0 {
		return true
	}
	if size <= user.GetRemainingCapacity() {
		user.Storage += size
		DB.Model(user).Update("storage", gorm.Expr("storage + ?", size))
		return true
	}
	return false
}

// ChangeStorage 更新用户容量
func (user *User) ChangeStorage(tx *gorm.DB, operator string, size uint64) error {
	return tx.Model(user).Update("storage", gorm.Expr("storage "+operator+" ?", size)).Error
}

// PayScore 扣除积分，返回是否成功
func (user *User) PayScore(score int) bool {
	if score == 0 {
		return true
	}
	if score <= user.Score {
		user.Score -= score
		DB.Model(user).Update("score", gorm.Expr("score - ?", score))
		return true
	}
	return false
}

// AddScore 增加积分
func (user *User) AddScore(score int) {
	user.Score += score
	DB.Model(user).Update("score", gorm.Expr("score + ?", score))
}

// IncreaseStorageWithoutCheck 忽略可用容量，增加用户已用容量
func (user *User) IncreaseStorageWithoutCheck(size uint64) {
	if size == 0 {
		return
	}
	user.Storage += size
	DB.Model(user).Update("storage", gorm.Expr("storage + ?", size))

}

// GetRemainingCapacity 获取剩余配额
func (user *User) GetRemainingCapacity() uint64 {
	total := user.Group.MaxStorage + user.GetAvailablePackSize()
	if total <= user.Storage {
		return 0
	}
	return total - user.Storage
}

// GetPolicyID 获取给定目录的存储策略, 如果为 nil 则使用默认
func (user *User) GetPolicyID(folder *Folder) *Policy {
	if user.IsAnonymous() {
		return &Policy{Type: "anonymous"}
	}

	defaultPolicy := uint(1)
	if len(user.Group.PolicyList) > 0 {
		defaultPolicy = user.Group.PolicyList[0]
	}

	if folder != nil {
		prefer := folder.PolicyID
		if prefer == 0 && folder.InheritPolicyID > 0 {
			prefer = folder.InheritPolicyID
		}

		if prefer > 0 && util.ContainsUint(user.Group.PolicyList, prefer) {
			defaultPolicy = prefer
		}
	}

	p, _ := GetPolicyByID(defaultPolicy)
	return &p
}

// GetPolicyByPreference 在可用存储策略中优先获取 preference
func (user *User) GetPolicyByPreference(preference uint) *Policy {
	if user.IsAnonymous() {
		return &Policy{Type: "anonymous"}
	}

	defaultPolicy := uint(1)
	if len(user.Group.PolicyList) > 0 {
		defaultPolicy = user.Group.PolicyList[0]
	}

	if preference != 0 {
		if util.ContainsUint(user.Group.PolicyList, preference) {
			defaultPolicy = preference
		}
	}

	p, _ := GetPolicyByID(defaultPolicy)
	return &p
}

// GetUserByID 用ID获取用户
func GetUserByID(ID interface{}) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).First(&user, ID)
	return user, result.Error
}

// GetActiveUserByID 用ID获取可登录用户
func GetActiveUserByID(ID interface{}) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).Where("status = ?", Active).First(&user, ID)
	return user, result.Error
}

// GetActiveUserByOpenID 用OpenID获取可登录用户
func GetActiveUserByOpenID(openid string) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).Where("status = ? and open_id = ?", Active, openid).Find(&user)
	return user, result.Error
}

// GetUserByEmail 用Email获取用户
func GetUserByEmail(email string) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).Where("email = ?", email).First(&user)
	return user, result.Error
}

// GetActiveUserByEmail 用Email获取可登录用户
func GetActiveUserByEmail(email string) (User, error) {
	var user User
	result := DB.Set("gorm:auto_preload", true).Where("status = ? and email = ?", Active, email).First(&user)
	return user, result.Error
}

// NewUser 返回一个新的空 User
func NewUser() User {
	options := UserOption{}
	return User{
		OptionsSerialized: options,
	}
}

// BeforeSave Save用户前的钩子
func (user *User) BeforeSave() (err error) {
	err = user.SerializeOptions()
	return err
}

// AfterCreate 创建用户后的钩子
func (user *User) AfterCreate(tx *gorm.DB) (err error) {
	// 创建用户的默认根目录
	defaultFolder := &Folder{
		Name:    "/",
		OwnerID: user.ID,
	}
	tx.Create(defaultFolder)

	// 创建用户初始文件记录
	initialFiles := GetSettingByNameFromTx(tx, "initial_files")
	if initialFiles != "" {
		initialFileIDs := make([]uint, 0)
		if err := json.Unmarshal([]byte(initialFiles), &initialFileIDs); err != nil {
			return err
		}

		if files, err := GetFilesByIDsFromTX(tx, initialFileIDs, 0); err == nil {
			for _, file := range files {
				file.ID = 0
				file.UserID = user.ID
				file.FolderID = defaultFolder.ID
				user.Storage += file.Size
				tx.Create(&file)
			}
			tx.Save(user)
		}
	}

	return err
}

// AfterFind 找到用户后的钩子
func (user *User) AfterFind() (err error) {
	// 解析用户设置到OptionsSerialized
	if user.Options != "" {
		err = json.Unmarshal([]byte(user.Options), &user.OptionsSerialized)
	}

	return err
}

// SerializeOptions 将序列后的Option写入到数据库字段
func (user *User) SerializeOptions() (err error) {
	optionsValue, err := json.Marshal(&user.OptionsSerialized)
	user.Options = string(optionsValue)
	return err
}

// CheckPassword 根据明文校验密码
func (user *User) CheckPassword(password string) (bool, error) {

	// 根据存储密码拆分为 Salt 和 Digest
	passwordStore := strings.Split(user.Password, ":")
	if len(passwordStore) != 2 && len(passwordStore) != 3 {
		return false, errors.New("Unknown password type")
	}

	// 兼容V2密码，升级后存储格式为: md5:$HASH:$SALT
	if len(passwordStore) == 3 {
		if passwordStore[0] != "md5" {
			return false, errors.New("Unknown password type")
		}
		hash := md5.New()
		_, err := hash.Write([]byte(passwordStore[2] + password))
		bs := hex.EncodeToString(hash.Sum(nil))
		if err != nil {
			return false, err
		}
		return bs == passwordStore[1], nil
	}

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

// NewAnonymousUser 返回一个匿名用户
func NewAnonymousUser() *User {
	user := User{}
	user.Group, _ = GetGroupByID(3)
	return &user
}

// IsAnonymous 返回是否为未登录用户
func (user *User) IsAnonymous() bool {
	return user.ID == 0
}

// Notified 更新用户容量超额通知日期
func (user *User) Notified() {
	if user.NotifyDate == nil {
		timeNow := time.Now()
		user.NotifyDate = &timeNow
		DB.Model(&user).Update("notify_date", user.NotifyDate)
	}
}

// ClearNotified 清除用户通知标记
func (user *User) ClearNotified() {
	DB.Model(&user).Update("notify_date", nil)
}

// SetStatus 设定用户状态
func (user *User) SetStatus(status int) {
	DB.Model(&user).Update("status", status)
}

// Update 更新用户
func (user *User) Update(val map[string]interface{}) error {
	return DB.Model(user).Updates(val).Error
}

// UpdateOptions 更新用户偏好设定
func (user *User) UpdateOptions() error {
	if err := user.SerializeOptions(); err != nil {
		return err
	}
	return user.Update(map[string]interface{}{"options": user.Options})
}

// GetGroupExpiredUsers 获取用户组过期的用户
func GetGroupExpiredUsers() []User {
	var users []User
	DB.Where("group_expires < ? and previous_group_id <> 0", time.Now()).Find(&users)
	return users
}

// GetTolerantExpiredUser 获取超过宽容期的用户
func GetTolerantExpiredUser() []User {
	var users []User
	DB.Set("gorm:auto_preload", true).Where("notify_date < ?", time.Now().Add(
		time.Duration(-GetIntSetting("ban_time", 10))*time.Second),
	).Find(&users)
	return users
}

// GroupFallback 回退到初始用户组
func (user *User) GroupFallback() {
	if user.GroupExpires != nil && user.PreviousGroupID != 0 {
		user.Group.ID = user.PreviousGroupID
		DB.Model(&user).Updates(map[string]interface{}{
			"group_expires":     nil,
			"previous_group_id": 0,
			"group_id":          user.PreviousGroupID,
		})
	}
}

// UpgradeGroup 升级用户组
func (user *User) UpgradeGroup(id uint, expires *time.Time) error {
	user.Group.ID = id
	previousGroupID := user.GroupID
	if user.PreviousGroupID != 0 && user.GroupID == id {
		previousGroupID = user.PreviousGroupID
	}
	return DB.Model(&user).Updates(map[string]interface{}{
		"group_expires":     expires,
		"previous_group_id": previousGroupID,
		"group_id":          id,
	}).Error
}
