package fsctx

import (
	"encoding/hex"
	"errors"
	"github.com/HFO4/aliyun-oss-go-sdk/oss"
	"hash"
	"io"
	"strings"
	"time"
)

type WriteMode int

const (
	Overwrite WriteMode = 0x00001
	// Append 只适用于本地策略
	Append WriteMode = 0x00002
	Nop    WriteMode = 0x00004
)

type UploadTaskInfo struct {
	Size            uint64
	MimeType        string
	FileName        string
	VirtualPath     string
	Mode            WriteMode
	Metadata        map[string]string
	LastModified    *time.Time
	SavePath        string
	UploadSessionID *string
	AppendStart     uint64
	Model           interface{}
	Src             string
}

// Get mimetype of uploaded file, if it's not defined, detect it from file name
func (u *UploadTaskInfo) DetectMimeType() string {
	if u.MimeType != "" {
		return u.MimeType
	}

	return oss.TypeByExtension(u.FileName)
}

// FileHeader 上传来的文件数据处理器
type FileHeader interface {
	io.Reader
	io.Closer
	io.Seeker
	Info() *UploadTaskInfo
	SetSize(uint64)
	SetModel(fileModel interface{})
	Seekable() bool
}

// FileStream 用户传来的文件
type FileStream struct {
	Mode            WriteMode
	LastModified    *time.Time
	Metadata        map[string]string
	File            io.ReadCloser
	Seeker          io.Seeker
	Size            uint64
	VirtualPath     string
	Name            string
	MimeType        string
	SavePath        string
	UploadSessionID *string
	AppendStart     uint64
	Model           interface{}
	Src             string
}

func (file *FileStream) Read(p []byte) (n int, err error) {
	if file.File != nil {
		return file.File.Read(p)
	}

	return 0, io.EOF
}

func (file *FileStream) Close() error {
	if file.File != nil {
		return file.File.Close()
	}

	return nil
}

func (file *FileStream) Seek(offset int64, whence int) (int64, error) {
	if file.Seekable() {
		return file.Seeker.Seek(offset, whence)
	}

	return 0, errors.New("no seeker")
}

func (file *FileStream) Seekable() bool {
	return file.Seeker != nil
}

func (file *FileStream) Info() *UploadTaskInfo {
	return &UploadTaskInfo{
		Size:            file.Size,
		MimeType:        file.MimeType,
		FileName:        file.Name,
		VirtualPath:     file.VirtualPath,
		Mode:            file.Mode,
		Metadata:        file.Metadata,
		LastModified:    file.LastModified,
		SavePath:        file.SavePath,
		UploadSessionID: file.UploadSessionID,
		AppendStart:     file.AppendStart,
		Model:           file.Model,
		Src:             file.Src,
	}
}

func (file *FileStream) SetSize(size uint64) {
	file.Size = size
}

func (file *FileStream) SetModel(fileModel interface{}) {
	file.Model = fileModel
}

type ChecksumFileStream struct {
	Md5  hash.Hash
	Sha1 hash.Hash
	io.Reader
	io.Closer
}

func (file *ChecksumFileStream) Hash() string {
	var sb strings.Builder

	sb.WriteString("MD5:")
	sb.WriteString(hex.EncodeToString(file.Md5.Sum(nil)))

	sb.WriteByte(' ')

	sb.WriteString("SHA1:")
	sb.WriteString(hex.EncodeToString(file.Sha1.Sum(nil)))

	return sb.String()
}
