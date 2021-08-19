package model

import "github.com/jinzhu/gorm"

// Node 从机节点信息模型
type Node struct {
	gorm.Model
	Status       NodeStatus // 节点状态
	Type         ModelType  // 节点状态
	Server       string     // 服务器地址
	SecretKey    string     `gorm:"type:text"` // 通信密钥
	Aria2Enabled bool       // 是否支持用作离线下载节点
	Aria2Options string     `gorm:"type:text"` // 离线下载配置
}

type NodeStatus int
type ModelType int

const (
	NodeActive = iota
	NodeSuspend
)

const (
	SlaveNodeType = iota
	MasterNodeType
)

// GetNodesByStatus 根据给定状态获取节点
func GetNodesByStatus(status ...NodeStatus) ([]Node, error) {
	var nodes []Node
	result := DB.Where("status in (?)", status).Find(&nodes)
	return nodes, result.Error
}
