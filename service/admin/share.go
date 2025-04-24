package admin

import (
	"context"
	"strconv"

	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/cluster/routes"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

const (
	shareUserIDCondition = "share_user_id"
	shareFileIDCondition = "share_file_id"
)

func (s *AdminListService) Shares(c *gin.Context) (*ListShareResponse, error) {
	dep := dependency.FromContext(c)
	shareClient := dep.ShareClient()
	hasher := dep.HashIDEncoder()

	var (
		err    error
		userID int
		fileID int
	)

	if s.Conditions[shareUserIDCondition] != "" {
		userID, err = strconv.Atoi(s.Conditions[shareUserIDCondition])
		if err != nil {
			return nil, serializer.NewError(serializer.CodeParamErr, "Invalid share user ID", err)
		}
	}

	if s.Conditions[shareFileIDCondition] != "" {
		fileID, err = strconv.Atoi(s.Conditions[shareFileIDCondition])
		if err != nil {
			return nil, serializer.NewError(serializer.CodeParamErr, "Invalid share file ID", err)
		}
	}

	ctx := context.WithValue(c, inventory.LoadShareFile{}, true)
	ctx = context.WithValue(ctx, inventory.LoadShareUser{}, true)

	res, err := shareClient.List(ctx, &inventory.ListShareArgs{
		PaginationArgs: &inventory.PaginationArgs{
			Page:     s.Page - 1,
			PageSize: s.PageSize,
			OrderBy:  s.OrderBy,
			Order:    inventory.OrderDirection(s.OrderDirection),
		},
		UserID: userID,
		FileID: fileID,
	})

	if err != nil {
		return nil, serializer.NewError(serializer.CodeDBError, "Failed to list shares", err)
	}

	siteUrl := dep.SettingProvider().SiteURL(c)

	return &ListShareResponse{
		Pagination: res.PaginationResults,
		Shares: lo.Map(res.Shares, func(share *ent.Share, _ int) GetShareResponse {
			var (
				uid       string
				shareLink string
			)

			if share.Edges.User != nil {
				uid = hashid.EncodeUserID(hasher, share.Edges.User.ID)
			}

			shareLink = routes.MasterShareUrl(siteUrl, hashid.EncodeShareID(hasher, share.ID), share.Password).String()

			return GetShareResponse{
				Share:      share,
				UserHashID: uid,
				ShareLink:  shareLink,
			}
		}),
	}, nil

}

type (
	SingleShareService struct {
		ShareID int `uri:"id" binding:"required"`
	}
	SingleShareParamCtx struct{}
)

func (s *SingleShareService) Get(c *gin.Context) (*GetShareResponse, error) {
	dep := dependency.FromContext(c)
	shareClient := dep.ShareClient()
	hasher := dep.HashIDEncoder()

	ctx := context.WithValue(c, inventory.LoadShareFile{}, true)
	ctx = context.WithValue(ctx, inventory.LoadShareUser{}, true)
	share, err := shareClient.GetByID(ctx, s.ShareID)
	if err != nil {
		return nil, serializer.NewError(serializer.CodeDBError, "Failed to get share", err)
	}

	var (
		uid       string
		shareLink string
	)

	if share.Edges.User != nil {
		uid = hashid.EncodeUserID(hasher, share.Edges.User.ID)
	}

	siteUrl := dep.SettingProvider().SiteURL(c)
	shareLink = routes.MasterShareUrl(siteUrl, hashid.EncodeShareID(hasher, share.ID), share.Password).String()

	return &GetShareResponse{
		Share:      share,
		UserHashID: uid,
		ShareLink:  shareLink,
	}, nil
}

type (
	BatchShareService struct {
		ShareIDs []int `json:"ids" binding:"required"`
	}
	BatchShareParamCtx struct{}
)

func (s *BatchShareService) Delete(c *gin.Context) error {
	dep := dependency.FromContext(c)
	shareClient := dep.ShareClient()

	if err := shareClient.DeleteBatch(c, s.ShareIDs); err != nil {
		return serializer.NewError(serializer.CodeDBError, "Failed to delete shares", err)
	}

	return nil
}
