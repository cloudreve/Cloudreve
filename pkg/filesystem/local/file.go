package local

import (
	"io"
	"mime/multipart"
)

// FileData 上传的文件数据
type FileData struct {
	File     multipart.File
	Size     uint64
	Name     string
	MIMEType string
}

func (file FileData) Read(p []byte) (n int, err error) {
	return file.File.Read(p)
}

func (file FileData) GetMIMEType() string {
	return file.MIMEType
}

func (file FileData) GetSize() uint64 {
	return file.Size
}

func (file FileData) Close() error {
	return file.File.Close()
}

func (file FileData) GetFileName() string {
	return file.Name
}

type FileStream struct {
	File     io.ReadCloser
	Size     uint64
	Name     string
	MIMEType string
}

func (file FileStream) Read(p []byte) (n int, err error) {
	return file.File.Read(p)
}

func (file FileStream) GetMIMEType() string {
	return file.MIMEType
}

func (file FileStream) GetSize() uint64 {
	return file.Size
}

func (file FileStream) Close() error {
	return file.File.Close()
}

func (file FileStream) GetFileName() string {
	return file.Name
}
