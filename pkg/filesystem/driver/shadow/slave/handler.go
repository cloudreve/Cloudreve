package slave

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"io"
	"net/url"
)

// Driver 影子存储策略，将上传任务指派给从机节点处理，并等待从机通知上传结果
type Driver struct {
	node    cluster.Node
	handler driver.Handler
	policy  *model.Policy
}

// NewDriver 返回新的从机指派处理器
func NewDriver(node cluster.Node, handler driver.Handler, policy *model.Policy) driver.Handler {
	return &Driver{
		node:    node,
		handler: handler,
		policy:  policy,
	}
}

func (d Driver) Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error {

	panic("implement me")
}

func (d Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	panic("implement me")
}

func (d Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	panic("implement me")
}

func (d Driver) Thumb(ctx context.Context, path string) (*response.ContentResponse, error) {
	panic("implement me")
}

func (d Driver) Source(ctx context.Context, path string, url url.URL, ttl int64, isDownload bool, speed int) (string, error) {
	panic("implement me")
}

func (d Driver) Token(ctx context.Context, ttl int64, callbackKey string) (serializer.UploadCredential, error) {
	panic("implement me")
}

func (d Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	panic("implement me")
}
