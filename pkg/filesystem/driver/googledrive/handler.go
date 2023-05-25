package googledrive

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

// Driver Google Drive 适配器
type Driver struct {
	Policy     *model.Policy
	HTTPClient request.Client
}

// NewDriver 从存储策略初始化新的Driver实例
func NewDriver(policy *model.Policy) (driver.Handler, error) {
	return &Driver{
		Policy:     policy,
		HTTPClient: request.NewClient(),
	}, nil
}

func (d *Driver) Put(ctx context.Context, file fsctx.FileHeader) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Delete(ctx context.Context, files []string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Get(ctx context.Context, path string) (response.RSCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Source(ctx context.Context, path string, ttl int64, isDownload bool, speed int) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) List(ctx context.Context, path string, recursive bool) ([]response.Object, error) {
	//TODO implement me
	panic("implement me")
}
