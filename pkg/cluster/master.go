package cluster

import (
	"context"
	"encoding/json"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gofrs/uuid"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	deleteTempFileDuration = 60 * time.Second
	statusRetryDuration    = 10 * time.Second
)

type MasterNode struct {
	Model    *model.Node
	aria2RPC rpcService
	lock     sync.RWMutex
}

// RPCService 通过RPC服务的Aria2任务管理器
type rpcService struct {
	Caller      rpc.Client
	Initialized bool

	retryDuration         time.Duration
	deletePaddingDuration time.Duration
	parent                *MasterNode
	options               *clientOptions
}

type clientOptions struct {
	Options map[string]interface{} // 创建下载时额外添加的设置
}

// Init 初始化节点
func (node *MasterNode) Init(nodeModel *model.Node) {
	node.lock.Lock()
	node.Model = nodeModel
	node.aria2RPC.parent = node
	node.aria2RPC.retryDuration = statusRetryDuration
	node.aria2RPC.deletePaddingDuration = deleteTempFileDuration
	node.lock.Unlock()

	node.lock.RLock()
	if node.Model.Aria2Enabled {
		node.lock.RUnlock()
		node.aria2RPC.Init()
		return
	}
	node.lock.RUnlock()
}

func (node *MasterNode) ID() uint {
	node.lock.RLock()
	defer node.lock.RUnlock()

	return node.Model.ID
}

func (node *MasterNode) Ping(req *serializer.NodePingReq) (*serializer.NodePingResp, error) {
	return &serializer.NodePingResp{}, nil
}

// IsFeatureEnabled 查询节点的某项功能是否启用
func (node *MasterNode) IsFeatureEnabled(feature string) bool {
	node.lock.RLock()
	defer node.lock.RUnlock()

	switch feature {
	case "aria2":
		return node.Model.Aria2Enabled
	default:
		return false
	}
}

func (node *MasterNode) MasterAuthInstance() auth.Auth {
	node.lock.RLock()
	defer node.lock.RUnlock()

	return auth.HMACAuth{SecretKey: []byte(node.Model.MasterKey)}
}

func (node *MasterNode) SlaveAuthInstance() auth.Auth {
	node.lock.RLock()
	defer node.lock.RUnlock()

	return auth.HMACAuth{SecretKey: []byte(node.Model.SlaveKey)}
}

// SubscribeStatusChange 订阅节点状态更改
func (node *MasterNode) SubscribeStatusChange(callback func(isActive bool, id uint)) {
}

// IsActive 返回节点是否在线
func (node *MasterNode) IsActive() bool {
	return true
}

// Kill 结束aria2请求
func (node *MasterNode) Kill() {
	if node.aria2RPC.Caller != nil {
		node.aria2RPC.Caller.Close()
	}
}

// GetAria2Instance 获取主机Aria2实例
func (node *MasterNode) GetAria2Instance() common.Aria2 {
	node.lock.RLock()

	if !node.Model.Aria2Enabled {
		node.lock.RUnlock()
		return &common.DummyAria2{}
	}

	if !node.aria2RPC.Initialized {
		node.lock.RUnlock()
		node.aria2RPC.Init()
		return &common.DummyAria2{}
	}

	defer node.lock.RUnlock()
	return &node.aria2RPC
}

func (node *MasterNode) IsMater() bool {
	return true
}

func (node *MasterNode) DBModel() *model.Node {
	node.lock.RLock()
	defer node.lock.RUnlock()

	return node.Model
}

func (r *rpcService) Init() error {
	r.parent.lock.Lock()
	defer r.parent.lock.Unlock()
	r.Initialized = false

	// 客户端已存在，则关闭先前连接
	if r.Caller != nil {
		r.Caller.Close()
	}

	// 解析RPC服务地址
	server, err := url.Parse(r.parent.Model.Aria2OptionsSerialized.Server)
	if err != nil {
		util.Log().Warning("无法解析主机 Aria2 RPC 服务地址，%s", err)
		return err
	}
	server.Path = "/jsonrpc"

	// 加载自定义下载配置
	var globalOptions map[string]interface{}
	if r.parent.Model.Aria2OptionsSerialized.Options != "" {
		err = json.Unmarshal([]byte(r.parent.Model.Aria2OptionsSerialized.Options), &globalOptions)
		if err != nil {
			util.Log().Warning("无法解析主机 Aria2 配置，%s", err)
			return err
		}
	}

	r.options = &clientOptions{
		Options: globalOptions,
	}
	timeout := r.parent.Model.Aria2OptionsSerialized.Timeout
	caller, err := rpc.New(context.Background(), server.String(), r.parent.Model.Aria2OptionsSerialized.Token, time.Duration(timeout)*time.Second, mq.GlobalMQ)

	r.Caller = caller
	r.Initialized = err == nil
	return err
}

func (r *rpcService) CreateTask(task *model.Download, groupOptions map[string]interface{}) (string, error) {
	r.parent.lock.RLock()
	// 生成存储路径
	guid, _ := uuid.NewV4()
	path := filepath.Join(
		r.parent.Model.Aria2OptionsSerialized.TempPath,
		"aria2",
		guid.String(),
	)
	r.parent.lock.RUnlock()

	// 创建下载任务
	options := map[string]interface{}{
		"dir": path,
	}
	for k, v := range r.options.Options {
		options[k] = v
	}
	for k, v := range groupOptions {
		options[k] = v
	}

	gid, err := r.Caller.AddURI(task.Source, options)
	if err != nil || gid == "" {
		return "", err
	}

	return gid, nil
}

func (r *rpcService) Status(task *model.Download) (rpc.StatusInfo, error) {
	res, err := r.Caller.TellStatus(task.GID)
	if err != nil {
		// 失败后重试
		util.Log().Debug("无法获取离线下载状态，%s，稍后重试", err)
		time.Sleep(r.retryDuration)
		res, err = r.Caller.TellStatus(task.GID)
	}

	return res, err
}

func (r *rpcService) Cancel(task *model.Download) error {
	// 取消下载任务
	_, err := r.Caller.Remove(task.GID)
	if err != nil {
		util.Log().Warning("无法取消离线下载任务[%s], %s", task.GID, err)
	}

	return err
}

func (r *rpcService) Select(task *model.Download, files []int) error {
	var selected = make([]string, len(files))
	for i := 0; i < len(files); i++ {
		selected[i] = strconv.Itoa(files[i])
	}
	_, err := r.Caller.ChangeOption(task.GID, map[string]interface{}{"select-file": strings.Join(selected, ",")})
	return err
}

func (r *rpcService) GetConfig() model.Aria2Option {
	r.parent.lock.RLock()
	defer r.parent.lock.RUnlock()

	return r.parent.Model.Aria2OptionsSerialized
}

func (s *rpcService) DeleteTempFile(task *model.Download) error {
	s.parent.lock.RLock()
	defer s.parent.lock.RUnlock()

	// 避免被aria2占用，异步执行删除
	go func(d time.Duration, src string) {
		time.Sleep(d)
		err := os.RemoveAll(src)
		if err != nil {
			util.Log().Warning("无法删除离线下载临时目录[%s], %s", src, err)
		}
	}(s.deletePaddingDuration, task.Parent)

	return nil
}
