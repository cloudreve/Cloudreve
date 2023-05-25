package masterinslave

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// Driver 影子存储策略，用于在从机端上传文件
type Driver struct {
	master  cluster.Node
	handler driver.Handler
	policy  *model.Policy
}

// NewDriver 返回新的处理器
func NewDriver(master cluster.Node, handler driver.Handler, policy *model.Policy) driver.Handler {
	return &Driver{
		master:  master,
		handler: handler,
		policy:  policy,
	}
}

func (d *Driver) Put(ctx context.Context, file fsctx.FileHeader) error {
	return d.handler.Put(ctx, file)
}

func (d *Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	return d.handler.Delete(ctx, files)
}

func (d *Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	return nil, ErrNotImplemented
}

func (d *Driver) Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error) {
	return nil, ErrNotImplemented
}

func (d *Driver) Source(ctx context.Context, path string, ttl int64, isDownload bool, speed int) (string, error) {
	return "", ErrNotImplemented
}

func (d *Driver) Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error) {
	return nil, ErrNotImplemented
}

func (d *Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	return nil, ErrNotImplemented
}

// 取消上传凭证
func (handler Driver) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	return nil
}
