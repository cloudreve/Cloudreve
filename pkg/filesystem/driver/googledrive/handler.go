package googledrive

import (
	"context"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"net/url"
)

// Driver Google Drive 适配器
type Driver struct {
	Policy     *model.Policy
	HTTPClient request.Client
}

// http://localhost:3000/api/v3/callback/googledrive/auth?code=4/0AVHEtk4AaNbo5YoCrMSGgoJfZfe6SgEOVmA7XtalZl8BMtdsAIRWqxt6jO4NKJCxGVxyQA&scope=profile%20openid%20https://www.googleapis.com/auth/userinfo.profile%20https://www.googleapis.com/auth/drive&authuser=0&prompt=consent
// https://accounts.google.com/o/oauth2/v2/auth?client_id=89866991293-5uja7qsbl8btuak3hb41h3o8u9jhlckg.apps.googleusercontent.com&response_type=code&redirect_uri=http://localhost:3000/api/v3/callback/googledrive/auth&scope=openid+profile+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive&access_type=offline&prompt=consent
//https://accounts.google.com/o/oauth2/auth?client_id=202264815644.apps.googleusercontent.com&response_type=code&redirect_uri=http%3A%2F%2F127.0.0.1%3A53682%2F&scope=openid+profile+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fphotoslibrary&access_type=offline&prompt=consent&state=MjAyMjY0ODE1NjQ0LmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29tOjpYNFozY2E4eGZXRGIxVm9vLUY5YTdaeEo6Omh0dHA6Ly8xMjcuMC4wLjE6NTM2ODIv

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

func (d *Driver) Source(ctx context.Context, path string, url url.URL, ttl int64, isDownload bool, speed int) (string, error) {
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
