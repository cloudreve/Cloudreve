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

	// Add given node into pool. If node existed, refresh node.
	Add(node *model.Node)

	// Delete and kill node from pool by given node id
	Delete(id uint)
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
	Default = &NodePool{}
	Default.Init()
	if err := Default.initFromDB(); err != nil {
		util.Log().Warning("节点池初始化失败, %s", err)
	}
}

func (pool *NodePool) Init() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.featureMap = make(map[string][]Node)
	pool.active = make(map[uint]Node)
	pool.inactive = make(map[uint]Node)
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
	var node Node
	pool.lock.Lock()
	if n, ok := pool.inactive[id]; ok {
		node = n
		delete(pool.inactive, id)
	} else {
		node = pool.active[id]
		delete(pool.active, id)
	}

	if isActive {
		pool.active[id] = node
	} else {
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
	for i := 0; i < len(nodes); i++ {
		pool.add(&nodes[i])
	}
	pool.lock.Unlock()

	pool.buildIndexMap()
	return nil
}

func (pool *NodePool) add(node *model.Node) {
	newNode := NewNodeFromDBModel(node)
	if newNode.IsActive() {
		pool.active[node.ID] = newNode
	} else {
		pool.inactive[node.ID] = newNode
	}

	// 订阅节点状态变更
	newNode.SubscribeStatusChange(func(isActive bool, id uint) {
		pool.nodeStatusChange(isActive, id)
	})
}

func (pool *NodePool) Add(node *model.Node) {
	pool.lock.Lock()
	defer pool.buildIndexMap()
	defer pool.lock.Unlock()

	var (
		old Node
		ok  bool
	)
	if old, ok = pool.active[node.ID]; !ok {
		old, ok = pool.inactive[node.ID]
	}
	if old != nil {
		go old.Init(node)
		return
	}

	pool.add(node)
}

func (pool *NodePool) Delete(id uint) {
	pool.lock.Lock()
	defer pool.buildIndexMap()
	defer pool.lock.Unlock()

	if node, ok := pool.active[id]; ok {
		node.Kill()
		delete(pool.active, id)
		return
	}

	if node, ok := pool.inactive[id]; ok {
		node.Kill()
		delete(pool.inactive, id)
		return
	}

}

// BalanceNodeByFeature 根据 feature 和 LoadBalancer 取出节点
func (pool *NodePool) BalanceNodeByFeature(feature string, lb balancer.Balancer) (error, Node) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	if nodes, ok := pool.featureMap[feature]; ok {
		err, res := lb.NextPeer(nodes)
		if err == nil {
			return nil, res.(Node)
		}

		return err, nil
	}

	return ErrFeatureNotExist, nil
}
