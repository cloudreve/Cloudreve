package mocks

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/common"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/balancer"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v3/pkg/task"
	testMock "github.com/stretchr/testify/mock"
	"io"
)

type NodePoolMock struct {
	testMock.Mock
}

func (n NodePoolMock) BalanceNodeByFeature(feature string, lb balancer.Balancer) (error, cluster.Node) {
	args := n.Called(feature, lb)
	return args.Error(0), args.Get(1).(cluster.Node)
}

func (n NodePoolMock) GetNodeByID(id uint) cluster.Node {
	args := n.Called(id)
	if res, ok := args.Get(0).(cluster.Node); ok {
		return res
	}

	return nil
}

func (n NodePoolMock) Add(node *model.Node) {
	n.Called(node)
}

func (n NodePoolMock) Delete(id uint) {
	n.Called(id)
}

type NodeMock struct {
	testMock.Mock
}

func (n NodeMock) Init(node *model.Node) {
	n.Called(node)
}

func (n NodeMock) IsFeatureEnabled(feature string) bool {
	args := n.Called(feature)
	return args.Bool(0)
}

func (n NodeMock) SubscribeStatusChange(callback func(isActive bool, id uint)) {
	n.Called(callback)
}

func (n NodeMock) Ping(req *serializer.NodePingReq) (*serializer.NodePingResp, error) {
	args := n.Called(req)
	return args.Get(0).(*serializer.NodePingResp), args.Error(1)
}

func (n NodeMock) IsActive() bool {
	args := n.Called()
	return args.Bool(0)
}

func (n NodeMock) GetAria2Instance() common.Aria2 {
	args := n.Called()
	return args.Get(0).(common.Aria2)
}

func (n NodeMock) ID() uint {
	args := n.Called()
	return args.Get(0).(uint)
}

func (n NodeMock) Kill() {
	n.Called()
}

func (n NodeMock) IsMater() bool {
	args := n.Called()
	return args.Bool(0)
}

func (n NodeMock) MasterAuthInstance() auth.Auth {
	args := n.Called()
	return args.Get(0).(auth.Auth)
}

func (n NodeMock) SlaveAuthInstance() auth.Auth {
	args := n.Called()
	return args.Get(0).(auth.Auth)
}

func (n NodeMock) DBModel() *model.Node {
	args := n.Called()
	return args.Get(0).(*model.Node)
}

type Aria2Mock struct {
	testMock.Mock
}

func (a Aria2Mock) Init() error {
	args := a.Called()
	return args.Error(0)
}

func (a Aria2Mock) CreateTask(task *model.Download, options map[string]interface{}) (string, error) {
	args := a.Called(task, options)
	return args.String(0), args.Error(1)
}

func (a Aria2Mock) Status(task *model.Download) (rpc.StatusInfo, error) {
	args := a.Called(task)
	return args.Get(0).(rpc.StatusInfo), args.Error(1)
}

func (a Aria2Mock) Cancel(task *model.Download) error {
	args := a.Called(task)
	return args.Error(0)
}

func (a Aria2Mock) Select(task *model.Download, files []int) error {
	args := a.Called(task, files)
	return args.Error(0)
}

func (a Aria2Mock) GetConfig() model.Aria2Option {
	args := a.Called()
	return args.Get(0).(model.Aria2Option)
}

func (a Aria2Mock) DeleteTempFile(download *model.Download) error {
	args := a.Called(download)
	return args.Error(0)
}

type TaskPoolMock struct {
	testMock.Mock
}

func (t TaskPoolMock) Add(num int) {
	t.Called(num)
}

func (t TaskPoolMock) Submit(job task.Job) {
	t.Called(job)
}

type RequestMock struct {
	testMock.Mock
}

func (r RequestMock) Request(method, target string, body io.Reader, opts ...request.Option) *request.Response {
	return r.Called(method, target, body, opts).Get(0).(*request.Response)
}
