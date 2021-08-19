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
	nodes      []Node
	featureMap map[string][]Node
	lock       sync.RWMutex
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

func (pool *NodePool) initFromDB() error {
	nodes, err := model.GetNodesByStatus(model.NodeActive)
	if err != nil {
		return err
	}

	pool.lock.Lock()

	for _, feature := range featureGroup {
		pool.featureMap[feature] = make([]Node, 0)
	}

	for i := 0; i < len(nodes); i++ {
		newNode := getNodeFromDBModel(&nodes[i])
		pool.nodes = append(pool.nodes, newNode)

		for _, feature := range featureGroup {
			if newNode.IsFeatureEnabled(feature) {
				pool.featureMap[feature] = append(pool.featureMap[feature], newNode)
			}
		}

		newNode.SubscribeStatusChange(func() {
		})
	}

	pool.lock.Unlock()
	return nil
}
