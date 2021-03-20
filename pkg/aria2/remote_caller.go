package aria2

import (
	"encoding/json"
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"net/url"
	"path"
	"strings"
)

// RemoteService 通过从机RPC服务的Aria2任务管理器
type RemoteService struct {
	Policy       *model.Policy
	Client       request.Client
	AuthInstance auth.Auth
}

func (client *RemoteService) Init(policy *model.Policy) {
	client.Policy = policy
	client.Client = request.HTTPClient{}
	client.AuthInstance = auth.HMACAuth{SecretKey: []byte(client.Policy.SecretKey)}
}

func (client *RemoteService) CreateTask(task *model.Download, options map[string]interface{}) error {
	reqBody := serializer.RemoteAria2AddRequest{
		TaskId:  task.ID,
		Options: options,
	}
	reqBodyEncoded, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// 发送列表请求
	bodyReader := strings.NewReader(string(reqBodyEncoded))
	signTTL := model.GetIntSetting("slave_api_timeout", 60)
	resp, err := client.Client.Request(
		"POST",
		client.getAPIUrl("add"),
		bodyReader,
		request.WithCredential(client.AuthInstance, int64(signTTL)),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return err
	}

	// 处理列取结果
	if resp.Code != 0 {
		return errors.New(resp.Error)
	}

	if resStr, ok := resp.Data.(string); ok {
		var res serializer.Response
		err = json.Unmarshal([]byte(resStr), &res)
		if err != nil {
			return err
		}
		if res.Code != 0 {
			return errors.New(res.Msg)
		}
	}

	return nil
}

func (client *RemoteService) Status(task *model.Download) (rpc.StatusInfo, error) {
	panic("implement me")
}

func (client *RemoteService) Cancel(task *model.Download) error {
	panic("implement me")
}

func (client *RemoteService) Select(task *model.Download, files []int) error {
	panic("implement me")
}

// getAPIUrl 获取接口请求地址
func (client *RemoteService) getAPIUrl(scope string, routes ...string) string {
	serverURL, err := url.Parse(client.Policy.Server)
	if err != nil {
		return ""
	}
	var controller *url.URL

	switch scope {
	case "add":
		controller, _ = url.Parse("/api/v3/slave/aria2/add")
	default:
		controller = serverURL
	}

	for _, r := range routes {
		controller.Path = path.Join(controller.Path, r)
	}

	return serverURL.ResolveReference(controller).String()
}
