package fsctx

import (
	"io"
	"time"
)

type WriteMode int

const (
	Overwrite WriteMode = iota
	// Append 只适用于本地策略
	Append
	Create
	Nop
)

// FileHeader 上传来的文件数据处理器
type FileHeader interface {
	io.Reader
	io.Closer
	Info() *UploadTaskInfo
	SetSize(uint64)
}

type UploadTaskInfo struct {
	Size            uint64
	MIMEType        string
	FileName        string
	VirtualPath     string
	Mode            WriteMode
	Metadata        map[string]string
	LastModified    *time.Time
	SavePath        string
	UploadSessionID *string
}

// FileStream 用户传来的文件
type FileStream struct {
	Mode            WriteMode
	LastModified    *time.Time
	Metadata        map[string]string
	File            io.ReadCloser
	Size            uint64
	VirtualPath     string
	Name            string
	MIMEType        string
	SavePath        string
	UploadSessionID *string
}

func (file *FileStream) Read(p []byte) (n int, err error) {
	return file.File.Read(p)
}

func (file *FileStream) Close() error {
	return file.File.Close()
}

func (file *FileStream) Info() *UploadTaskInfo {
	return &UploadTaskInfo{
		Size:            file.Size,
		MIMEType:        file.MIMEType,
		FileName:        file.Name,
		VirtualPath:     file.VirtualPath,
		Mode:            file.Mode,
		Metadata:        file.Metadata,
		LastModified:    file.LastModified,
		SavePath:        file.SavePath,
		UploadSessionID: file.UploadSessionID,
	}
}

func (file *FileStream) SetSize(size uint64) {
	file.Size = size
}
