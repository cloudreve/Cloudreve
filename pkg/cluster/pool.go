package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"sync"
)

var Default *NodePool

// 需要分类的节点组
var featureGroup = []string{"Aria2"}

// Pool 节点池
type Pool interface {
	Select()
}

// NodePool 通用节点池
type NodePool struct {
	active   map[uint]Node
	inactive map[uint]Node

	featureMap map[string][]Node

	lock sync.RWMutex
}

func (pool *NodePool) Select() {

}

// Init 初始化从机节点池
func Init() {
	Default = &NodePool{
		featureMap: make(map[string][]Node),
	}
	if err := Default.initFromDB(); err != nil {
		util.Log().Warning("节点池初始化失败, %s", err)
	}
}

func (pool *NodePool) buildIndexMap() {
	pool.lock.Lock()
	for _, feature := range featureGroup {
		pool.featureMap[feature] = make([]Node, 0)
	}

	for _, v := range pool.active {
		for _, feature := range featureGroup {
			if v.IsFeatureEnabled(feature) {
				pool.featureMap[feature] = append(pool.featureMap[feature], v)
			}
		}
	}
	pool.lock.Unlock()
}

func (pool *NodePool) nodeStatusChange(isActive bool, id uint) {
	util.Log().Debug("从机节点 [ID=%d] 状态变更 [active=%t]", id, isActive)
	pool.lock.Lock()
	if isActive {
		node := pool.inactive[id]
		delete(pool.inactive, id)
		pool.active[id] = node
	} else {
		node := pool.active[id]
		delete(pool.active, id)
		pool.inactive[id] = node
	}
	pool.lock.Unlock()

	pool.buildIndexMap()
}

func (pool *NodePool) initFromDB() error {
	nodes, err := model.GetNodesByStatus(model.NodeActive)
	if err != nil {
		return err
	}

	pool.lock.Lock()
	pool.active = make(map[uint]Node)
	pool.inactive = make(map[uint]Node)
	for i := 0; i < len(nodes); i++ {
		newNode := getNodeFromDBModel(&nodes[i])
		if newNode.IsActive() {
			pool.active[nodes[i].ID] = newNode
		} else {
			pool.inactive[nodes[i].ID] = newNode
		}

		// 订阅节点状态变更
		newNode.SubscribeStatusChange(func(isActive bool, id uint) {
			pool.nodeStatusChange(isActive, id)
		})
	}
	pool.lock.Unlock()

	pool.buildIndexMap()
	return nil
}
