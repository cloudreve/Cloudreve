package fsctx

import (
	"io"
	"time"
)

type WriteMode int

const (
	Overwrite WriteMode = iota
	Append
	Create
)

// FileStream 用户传来的文件
type FileStream struct {
	Mode         WriteMode
	Hidden       bool
	LastModified time.Time
	Metadata     map[string]string
	File         io.ReadCloser
	Size         uint64
	VirtualPath  string
	Name         string
	MIMEType     string
	SavePath     string
}

func (file *FileStream) Read(p []byte) (n int, err error) {
	return file.File.Read(p)
}

func (file *FileStream) GetMIMEType() string {
	return file.MIMEType
}

func (file *FileStream) GetSize() uint64 {
	return file.Size
}

func (file *FileStream) Close() error {
	return file.File.Close()
}

func (file *FileStream) GetFileName() string {
	return file.Name
}

func (file *FileStream) GetVirtualPath() string {
	return file.VirtualPath
}

func (file *FileStream) GetMode() WriteMode {
	return file.Mode
}

func (file *FileStream) GetMetadata() map[string]string {
	return file.Metadata
}

func (file *FileStream) GetLastModified() time.Time {
	return file.LastModified
}

func (file *FileStream) IsHidden() bool {
	return file.Hidden
}

func (file *FileStream) GetSavePath() string {
	return file.SavePath
}

// FileHeader 上传来的文件数据处理器
type FileHeader interface {
	io.Reader
	io.Closer
	GetSize() uint64
	GetMIMEType() string
	GetFileName() string
	GetVirtualPath() string
	GetMode() WriteMode
	GetMetadata() map[string]string
	GetLastModified() time.Time
	IsHidden() bool
	GetSavePath() string
}
