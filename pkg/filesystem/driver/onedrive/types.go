package onedrive

import (
	"io"
	"sync"
)

// RespError 接口返回错误
type RespError struct {
	APIError APIError `json:"error"`
}

// APIError 接口返回的错误内容
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UploadSessionResponse 分片上传会话
type UploadSessionResponse struct {
	DataContext        string   `json:"@odata.context"`
	ExpirationDateTime string   `json:"expirationDateTime"`
	NextExpectedRanges []string `json:"nextExpectedRanges"`
	UploadURL          string   `json:"uploadUrl"`
}

// FileInfo 文件元信息
type FileInfo struct {
	Name            string          `json:"name"`
	Size            uint64          `json:"size"`
	Image           imageInfo       `json:"image"`
	ParentReference parentReference `json:"parentReference"`
	DownloadURL     string          `json:"@microsoft.graph.downloadUrl"`
}

type imageInfo struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type parentReference struct {
	Path string `json:"path"`
	Name string `json:"name"`
	ID   string `json:"id"`
}

// UploadResult 上传结果
type UploadResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

// BatchRequests 批量操作请求
type BatchRequests struct {
	Requests []BatchRequest `json:"requests"`
}

// BatchRequest 批量操作单个请求
type BatchRequest struct {
	ID      string            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    interface{}       `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// BatchResponses 批量操作响应
type BatchResponses struct {
	Responses []BatchResponse `json:"responses"`
}

// BatchResponse 批量操作单个响应
type BatchResponse struct {
	ID     string `json:"id"`
	Status int    `json:"status"`
}

// ThumbResponse 获取缩略图的响应
type ThumbResponse struct {
	Value []map[string]interface{} `json:"value"`
}

// Chunk 文件分片
type Chunk struct {
	Offset    int
	ChunkSize int
	Total     int
	Retried   int
	Reader    io.Reader
}

// IsLast 返回是否为最后一个分片
func (chunk *Chunk) IsLast() bool {
	return chunk.Total-chunk.Offset == chunk.ChunkSize
}

var callbackSignal sync.Map
