package driver

import (
	"context"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/response"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"io"
	"net/url"
)

// Handler 存储策略适配器
type Handler interface {
	// 上传文件, dst为文件存储路径，size 为文件大小。上下文关闭
	// 时，应取消上传并清理临时文件
	Put(ctx context.Context, file io.ReadCloser, dst string, size uint64) error

	// 删除一个或多个给定路径的文件，返回删除失败的文件路径列表及错误
	Delete(ctx context.Context, files []string) ([]string, error)

	// 获取文件内容
	Get(ctx context.Context, path string) (response.RSCloser, error)

	// 获取缩略图，可直接在ContentResponse中返回文件数据流，也可指
	// 定为重定向
	Thumb(ctx context.Context, path string) (*response.ContentResponse, error)

	// 获取外链/下载地址，
	// url - 站点本身地址,
	// isDownload - 是否直接下载
	Source(ctx context.Context, path string, url url.URL, ttl int64, isDownload bool, speed int) (string, error)

	// Token 获取有效期为ttl的上传凭证和签名，同时回调会话有效期为sessionTTL
	Token(ctx context.Context, ttl int64, callbackKey string) (serializer.UploadCredential, error)

	// List 递归列取远程端path路径下文件、目录，不包含path本身，
	// 返回的对象路径以path作为起始根目录.
	// recursive - 是否递归列出
	List(ctx context.Context, path string, recursive bool) ([]response.Object, error)
}
