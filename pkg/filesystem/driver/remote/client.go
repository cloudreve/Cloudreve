package remote

import (
	"context"
	"encoding/json"
	"fmt"
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/auth"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/request"
	"github.com/cloudreve/Cloudreve/v3/pkg/retry"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/gofrs/uuid"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	basePath        = "/api/v3/slave/"
	OverwriteHeader = auth.CrHeaderPrefix + "Overwrite"
	chunkRetrySleep = time.Duration(5) * time.Second
)

// Client to operate remote slave server
type Client interface {
	// CreateUploadSession creates remote upload session
	CreateUploadSession(ctx context.Context, session *serializer.UploadSession, ttl int64) error
	// GetUploadURL signs an url for uploading file
	GetUploadURL(ttl int64, sessionID string) (string, string, error)
	// Upload uploads file to remote server
	Upload(ctx context.Context, file fsctx.FileHeader) error
}

// NewClient creates new Client from given policy
func NewClient(policy *model.Policy) (Client, error) {
	authInstance := auth.HMACAuth{[]byte(policy.SecretKey)}
	serverURL, err := url.Parse(policy.Server)
	if err != nil {
		return nil, err
	}

	base, _ := url.Parse(basePath)
	signTTL := model.GetIntSetting("slave_api_timeout", 60)

	return &remoteClient{
		policy:       policy,
		authInstance: authInstance,
		httpClient: request.NewClient(
			request.WithEndpoint(serverURL.ResolveReference(base).String()),
			request.WithCredential(authInstance, int64(signTTL)),
			request.WithMasterMeta(),
		),
	}, nil
}

type remoteClient struct {
	policy       *model.Policy
	authInstance auth.Auth
	httpClient   request.Client
}

func (c *remoteClient) Upload(ctx context.Context, file fsctx.FileHeader) error {
	ttl := model.GetIntSetting("upload_session_timeout", 86400)
	fileInfo := file.Info()
	session := &serializer.UploadSession{
		Key:          uuid.Must(uuid.NewV4()).String(),
		VirtualPath:  fileInfo.VirtualPath,
		Name:         fileInfo.FileName,
		Size:         fileInfo.Size,
		SavePath:     fileInfo.SavePath,
		LastModified: fileInfo.LastModified,
		Policy:       *c.policy,
	}

	// Create upload session
	if err := c.CreateUploadSession(ctx, session, int64(ttl)); err != nil {
		return fmt.Errorf("failed to create upload session: %w", err)
	}

	overwrite := fileInfo.Mode&fsctx.Overwrite == fsctx.Overwrite

	// Upload chunks
	offset := uint64(0)
	chunkSize := session.Policy.OptionsSerialized.ChunkSize
	if chunkSize == 0 {
		chunkSize = fileInfo.Size
	}

	chunkNum := fileInfo.Size / chunkSize
	if fileInfo.Size%chunkSize != 0 {
		chunkNum++
	}

	for i := 0; i < int(chunkNum); i++ {
		uploadFunc := func(index int, chunk io.Reader) error {
			contentLength := chunkSize
			if index == int(chunkNum-1) {
				contentLength = fileInfo.Size - chunkSize*(chunkNum-1)
			}

			return c.uploadChunk(ctx, session.Key, index, chunk, overwrite, contentLength)
		}

		if err := retry.Chunk(i, chunkSize, file, uploadFunc, retry.ConstantBackoff{
			Max:   model.GetIntSetting("onedrive_chunk_retries", 1),
			Sleep: chunkRetrySleep,
		}); err != nil {
			// TODO 删除上传会话
			return fmt.Errorf("failed to upload chunk #%d: %w", i, err)
		}

		offset += chunkSize
	}

	return nil
}

func (c *remoteClient) CreateUploadSession(ctx context.Context, session *serializer.UploadSession, ttl int64) error {
	reqBodyEncoded, err := json.Marshal(map[string]interface{}{
		"session": session,
		"ttl":     ttl,
	})
	if err != nil {
		return err
	}

	bodyReader := strings.NewReader(string(reqBodyEncoded))
	resp, err := c.httpClient.Request(
		"PUT",
		"upload",
		bodyReader,
		request.WithContext(ctx),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return serializer.NewErrorFromResponse(resp)
	}

	return nil
}

func (c *remoteClient) GetUploadURL(ttl int64, sessionID string) (string, string, error) {
	base, err := url.Parse(c.policy.Server)
	if err != nil {
		return "", "", err
	}

	base.Path = path.Join(base.Path, basePath, "upload", sessionID)
	req, err := http.NewRequest("POST", base.String(), nil)
	if err != nil {
		return "", "", err
	}

	req = auth.SignRequest(c.authInstance, req, ttl)
	return req.URL.String(), req.Header["Authorization"][0], nil
}

func (c *remoteClient) uploadChunk(ctx context.Context, sessionID string, index int, chunk io.Reader, overwrite bool, size uint64) error {
	resp, err := c.httpClient.Request(
		"POST",
		fmt.Sprintf("upload/%s?chunk=%d", sessionID, index),
		io.LimitReader(chunk, int64(size)),
		request.WithContext(ctx),
		request.WithTimeout(time.Duration(0)),
		request.WithContentLength(int64(size)),
		request.WithHeader(map[string][]string{OverwriteHeader: {fmt.Sprintf("%t", overwrite)}}),
	).CheckHTTPResponse(200).DecodeResponse()
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return serializer.NewErrorFromResponse(resp)
	}

	return nil
}
