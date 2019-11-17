package local

import "mime/multipart"

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
