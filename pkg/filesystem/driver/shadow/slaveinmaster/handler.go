package slaveinmaster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/mq"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"io"
	"net/url"
	"time"
)

// Driver 影子存储策略，将上传任务指派给从机节点处理，并等待从机通知上传结果
type Driver struct {
	node    cluster.Node
	handler driver.Handler
	policy  *model.Policy
	client  request.Client
}

// NewDriver 返回新的从机指派处理器
func NewDriver(node cluster.Node, handler driver.Handler, policy *model.Policy) driver.Handler {
	var endpoint *url.URL
	if serverURL, err := url.Parse(node.DBModel().Server); err == nil {
		var controller *url.URL
		controller, _ = url.Parse("/api/v3/slave")
		endpoint = serverURL.ResolveReference(controller)
	}

	signTTL := model.GetIntSetting("slave_api_timeout", 60)
	return &Driver{
		node:    node,
		handler: handler,
		policy:  policy,
		client: request.NewClient(
			request.WithMasterMeta(),
			request.WithTimeout(time.Duration(signTTL)*time.Second),
			request.WithCredential(node.SlaveAuthInstance(), int64(signTTL)),
			request.WithEndpoint(endpoint.String()),
		),
	}
}

// Put 将ctx中指定的从机物理文件由从机上传到目标存储策略
func (d *Driver) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {
	src, ok := ctx.Value(fsctx.SlaveSrcPath).(string)
	if !ok {
		return ErrSlaveSrcPathNotExist
	}

	req := serializer.SlaveTransferReq{
		Src:    src,
		Dst:    dst,
		Policy: d.policy,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// 订阅转存结果
	resChan := mq.GlobalMQ.Subscribe(req.Hash(model.GetSettingByName("siteID")), 0)
	defer mq.GlobalMQ.Unsubscribe(req.Hash(model.GetSettingByName("siteID")), resChan)

	res, err := d.client.Request("PUT", "task/transfer", bytes.NewReader(body)).
		CheckHTTPResponse(200).
		DecodeResponse()
	if err != nil {
		return err
	}

	if res.Code != 0 {
		return serializer.NewErrorFromResponse(res)
	}

	// 等待转存结果或者超时
	waitTimeout := model.GetIntSetting("slave_transfer_timeout", 172800)
	select {
	case <-time.After(time.Duration(waitTimeout) * time.Second):
		return ErrWaitResultTimeout
	case msg := <-resChan:
		if msg.Event != serializer.SlaveTransferSuccess {
			return errors.New(msg.Content.(serializer.SlaveTransferResult).Error)
		}
	}

	return nil
}

func (d *Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	return d.handler.Delete(ctx, files)
}

func (d *Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	return nil, ErrNotImplemented
}

func (d *Driver) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	return nil, ErrNotImplemented
}

func (d *Driver) Source(ctx context.Context, path string, url url.URL, ttl int64, isDownload bool, speed int) (string, error) {
	return "", ErrNotImplemented
}

func (d *Driver) Token(ctx context.Context, ttl int64, callbackKey string) (serializer.UploadCredential, error) {
	return serializer.UploadCredential{}, ErrNotImplemented
}

func (d *Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	return nil, ErrNotImplemented
}
