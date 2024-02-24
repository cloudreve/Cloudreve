package remoteclientmock

import (
	"context"

	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/serializer"
	"github.com/stretchr/testify/mock"
)

type RemoteClientMock struct {
	mock.Mock
}

func (r *RemoteClientMock) CreateUploadSession(ctx context.Context, session *serializer.UploadSession, ttl int64, overwrite bool) error {
	return r.Called(ctx, session, ttl, overwrite).Error(0)
}

func (r *RemoteClientMock) GetUploadURL(ttl int64, sessionID string) (string, string, error) {
	args := r.Called(ttl, sessionID)

	return args.String(0), args.String(1), args.Error(2)
}

func (r *RemoteClientMock) Upload(ctx context.Context, file fsctx.FileHeader) error {
	args := r.Called(ctx, file)
	return args.Error(0)
}

func (r *RemoteClientMock) DeleteUploadSession(ctx context.Context, sessionID string) error {
	args := r.Called(ctx, sessionID)
	return args.Error(0)
}
