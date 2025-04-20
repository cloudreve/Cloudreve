package model

import (
	"github.com/jinzhu/gorm"
)

// Node 从机节点信息模型
type Node struct {
	gorm.Model
	Status       NodeStatus // 节点状态
	Name         string     // 节点别名
	Type         ModelType  // 节点状态
	Server       string     // 服务器地址
	SlaveKey     string     `gorm:"type:text"` // 主->从 通信密钥
	MasterKey    string     `gorm:"type:text"` // 从->主 通信密钥
	Aria2Enabled bool       // 是否支持用作离线下载节点
	Aria2Options string     `gorm:"type:text"` // 离线下载配置
	Rank         int        // 负载均衡权重

	// 数据库忽略字段
	Aria2OptionsSerialized Aria2Option `gorm:"-"`
}

// Aria2Option 非公有的Aria2配置属性
type Aria2Option struct {
	// RPC 服务器地址
	Server string `json:"server,omitempty"`
	// RPC 密钥
	Token string `json:"token,omitempty"`
	// 临时下载目录
	TempPath string `json:"temp_path,omitempty"`
	// 附加下载配置
	Options string `json:"options,omitempty"`
	// 下载监控间隔
	Interval int `json:"interval,omitempty"`
	// RPC API 请求超时
	Timeout int `json:"timeout,omitempty"`
}

type NodeStatus int
type ModelType int

const (
	NodeActive NodeStatus = iota
	NodeSuspend
)

const (
	SlaveNodeType ModelType = iota
	MasterNodeType
)
