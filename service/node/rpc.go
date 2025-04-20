package node

import (
	"context"
	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/credmanager"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/manager"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/gin-gonic/gin"
)

type SlaveNotificationService struct {
	Subject string `uri:"subject" binding:"required"`
}

type (
	OauthCredentialParamCtx struct{}
	OauthCredentialService  struct {
		ID string `uri:"id" binding:"required"`
	}
)

// Get 获取主机Oauth策略的AccessToken
func (s *OauthCredentialService) Get(c *gin.Context) (*credmanager.CredentialResponse, error) {
	dep := dependency.FromContext(c)
	credManager := dep.CredManager()

	cred, err := credManager.Obtain(c, s.ID)
	if cred == nil || err != nil {
		return nil, serializer.NewError(serializer.CodeNotFound, "Credential not found", err)
	}

	return &credmanager.CredentialResponse{
		Token:    cred.String(),
		ExpireAt: cred.Expiry(),
	}, nil
}

type (
	StatelessPrepareUploadParamCtx struct{}
)

func StatelessPrepareUpload(s *fs.StatelessPrepareUploadService, c *gin.Context) (*fs.StatelessPrepareUploadResponse, error) {
	dep := dependency.FromContext(c)
	userClient := dep.UserClient()
	user, err := userClient.GetLoginUserByID(c, s.UserID)
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(c.Request.Context(), inventory.UserCtx{}, user)
	fm := manager.NewFileManager(dep, user)
	uploadSession, err := fm.PrepareUpload(ctx, s.UploadRequest)
	if err != nil {
		return nil, err
	}
	return &fs.StatelessPrepareUploadResponse{
		Session: uploadSession,
		Req:     s.UploadRequest,
	}, nil
}

type (
	StatelessCompleteUploadParamCtx struct{}
)

func StatelessCompleteUpload(s *fs.StatelessCompleteUploadService, c *gin.Context) (fs.File, error) {
	dep := dependency.FromContext(c)
	userClient := dep.UserClient()
	user, err := userClient.GetLoginUserByID(c, s.UserID)
	if err != nil {
		return nil, err
	}

	util.WithValue(c, inventory.UserCtx{}, user)
	fm := manager.NewFileManager(dep, user)
	return fm.CompleteUpload(c, s.UploadSession)
}

type (
	StatelessOnUploadFailedParamCtx struct{}
)

func StatelessOnUploadFailed(s *fs.StatelessOnUploadFailedService, c *gin.Context) error {
	dep := dependency.FromContext(c)
	userClient := dep.UserClient()
	user, err := userClient.GetLoginUserByID(c, s.UserID)
	if err != nil {
		return err
	}

	util.WithValue(c, inventory.UserCtx{}, user)
	fm := manager.NewFileManager(dep, user)
	fm.OnUploadFailed(c, s.UploadSession)
	return nil
}

type StatelessCreateFileParamCtx struct{}

func StatelessCreateFile(s *fs.StatelessCreateFileService, c *gin.Context) error {
	dep := dependency.FromContext(c)
	userClient := dep.UserClient()
	user, err := userClient.GetLoginUserByID(c, s.UserID)
	if err != nil {
		return err
	}

	uri, err := fs.NewUriFromString(s.Path)
	if err != nil {
		return err
	}

	util.WithValue(c, inventory.UserCtx{}, user)
	fm := manager.NewFileManager(dep, user)
	_, err = fm.Create(c, uri, s.Type)
	return err
}
