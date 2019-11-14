package model

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
)

// Policy 存储策略
type Policy struct {
	// 表字段
	gorm.Model
	Name               string
	Type               string
	Server             string
	BucketName         string
	IsPrivate          bool
	BaseURL            string
	AccessKey          string `gorm:"size:512"`
	SecretKey          string `gorm:"size:512"`
	MaxSize            uint64
	AutoRename         bool
	DirNameRule        string
	FileNameRule       string
	IsOriginLinkEnable bool
	Options            string `gorm:"size:4096"`

	// 数据库忽略字段
	OptionsSerialized PolicyOption `gorm:"-"`
}

// PolicyOption 非公有的存储策略属性
type PolicyOption struct {
	OPName               string   `json:"op_name"`
	OPPassword           string   `json:"op_pwd"`
	FileType             []string `json:"file_type"`
	MimeType             string   `json:"mimetype"`
	SpeedLimit           int      `json:"speed_limit"`
	RangeTransferEnabled bool     `json:"range_transfer_enabled"`
}

// GetPolicyByID 用ID获取存储策略
func GetPolicyByID(ID interface{}) (Policy, error) {
	var policy Policy
	result := DB.First(&policy, ID)
	return policy, result.Error
}

// AfterFind 找到上传策略后的钩子
func (policy *Policy) AfterFind() (err error) {
	// 解析上传策略设置到OptionsSerialized
	err = json.Unmarshal([]byte(policy.Options), &policy.OptionsSerialized)
	return err
}

// BeforeSave Save策略前的钩子
func (policy *Policy) BeforeSave() (err error) {
	err = policy.SerializeOptions()
	return err
}

//SerializeOptions 将序列后的Option写入到数据库字段
func (policy *Policy) SerializeOptions() (err error) {
	optionsValue, err := json.Marshal(&policy.OptionsSerialized)
	policy.Options = string(optionsValue)
	return err
}
