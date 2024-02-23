package driver

import (
	"context"
	"fmt"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
)

var (
	ErrorThumbNotExist     = fmt.Errorf("thumb not exist")
	ErrorThumbNotSupported = fmt.Errorf("thumb not supported")
)

// Handler 存储策略适配器
type Handler interface {
	// 上传文件, dst为文件存储路径，size 为文件大小。上下文关闭
	// 时，应取消上传并清理临时文件
	Put(ctx context.Context, file fsctx.FileHeader) error

	// 删除一个或多个给定路径的文件，返回删除失败的文件路径列表及错误
	Delete(ctx context.Context, files []string) ([]string, error)

	// 获取文件内容
	Get(ctx context.Context, path string) (response.RSCloser, error)

	// 获取缩略图，可直接在ContentResponse中返回文件数据流，也可指
	// 定为重定向
	// 	如果缩略图不存在, 且需要 Cloudreve 代理生成并上传，应返回 ErrorThumbNotExist，生
	//  成的缩略图文件存储规则与本机策略一致。
	// 	如果不支持此文件的缩略图，并且不希望后续继续请求此缩略图，应返回 ErrorThumbNotSupported
	Thumb(ctx context.Context, file *model.File) (*response.ContentResponse, error)

	// 获取外链/下载地址，
	// url - 站点本身地址,
	// isDownload - 是否直接下载
	Source(ctx context.Context, path string, ttl int64, isDownload bool, speed int) (string, error)

	// Token 获取有效期为ttl的上传凭证和签名
	Token(ctx context.Context, ttl int64, uploadSession *serializer.UploadSession, file fsctx.FileHeader) (*serializer.UploadCredential, error)

	// CancelToken 取消已经创建的有状态上传凭证
	CancelToken(ctx context.Context, uploadSession *serializer.UploadSession) error

	// List 递归列取远程端path路径下文件、目录，不包含path本身，
	// 返回的对象路径以path作为起始根目录.
	// recursive - 是否递归列出
	List(ctx context.Context, path string, recursive bool) ([]response.Object, error)
}
