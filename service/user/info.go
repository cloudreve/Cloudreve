package user

import (
	"context"
	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/manager"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func GetUser(c *gin.Context) (*ent.User, error) {
	uid := hashid.FromContext(c)
	dep := dependency.FromContext(c)
	userClient := dep.UserClient()
	ctx := context.WithValue(c, inventory.LoadUserGroup{}, true)
	return userClient.GetByID(ctx, uid)
}

func GetUserCapacity(c *gin.Context) (*fs.Capacity, error) {
	user := inventory.UserFromContext(c)
	dep := dependency.FromContext(c)
	m := manager.NewFileManager(dep, user)
	defer m.Recycle()

	return m.Capacity(c)
}

type (
	SearchUserService struct {
		Keyword string `form:"keyword" binding:"required,min=2"`
	}
	SearchUserParamCtx struct{}
)

const resultLimit = 10

func (s *SearchUserService) Search(c *gin.Context) ([]*ent.User, error) {
	dep := dependency.FromContext(c)
	userClient := dep.UserClient()
	res, err := userClient.SearchActive(c, resultLimit, s.Keyword)
	if err != nil {
		return nil, serializer.NewError(serializer.CodeDBError, "Failed to search user", err)
	}

	return res, nil
}

// ListAllGroups lists all groups.
func ListAllGroups(c *gin.Context) ([]*ent.Group, error) {
	dep := dependency.FromContext(c)
	groupClient := dep.GroupClient()
	res, err := groupClient.ListAll(c)
	if err != nil {
		return nil, serializer.NewError(serializer.CodeDBError, "Failed to list all groups", err)
	}

	res = lo.Filter(res, func(g *ent.Group, index int) bool {
		return g.ID != inventory.AnonymousGroupID
	})

	return res, nil
}
