package cluster

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/balancer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"sync"
)

var Default *NodePool

// 需要分类的节点组
var featureGroup = []string{"aria2"}

// Pool 节点池
type Pool interface {
	// Returns active node selected by given feature and load balancer
	BalanceNodeByFeature(feature string, lb balancer.Balancer) (error, Node)

	// Returns node by ID
	GetNodeByID(id uint) Node
}

// NodePool 通用节点池
type NodePool struct {
	active   map[uint]Node
	inactive map[uint]Node

	featureMap map[string][]Node

	lock sync.RWMutex
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

func (pool *NodePool) GetNodeByID(id uint) Node {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if node, ok := pool.active[id]; ok {
		return node
	}

	return pool.inactive[id]
}

func (pool *NodePool) nodeStatusChange(isActive bool, id uint) {
	util.Log().Debug("从机节点 [ID=%d] 状态变更 [Active=%t]", id, isActive)
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
		newNode := NewNodeFromDBModel(&nodes[i])
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

// BalanceNodeByFeature 根据 feature 和 LoadBalancer 取出节点
func (pool *NodePool) BalanceNodeByFeature(feature string, lb balancer.Balancer) (error, Node) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	if nodes, ok := pool.featureMap[feature]; ok {
		err, res := lb.NextPeer(nodes)
		return err, res.(Node)
	}

	return ErrFeatureNotExist, nil
}
