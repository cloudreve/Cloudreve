package serializer

import (
	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"time"
)

type quota struct {
	Base  uint64         `json:"base"`
	Pack  uint64         `json:"pack"`
	Used  uint64         `json:"used"`
	Total uint64         `json:"total"`
	Packs []storagePacks `json:"packs"`
}

type storagePacks struct {
	Name           string    `json:"name"`
	Size           uint64    `json:"size"`
	ActivateDate   time.Time `json:"activate_date"`
	Expiration     int       `json:"expiration"`
	ExpirationDate time.Time `json:"expiration_date"`
}

// MountedFolders 已挂载的目录
type MountedFolders struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PolicyName string `json:"policy_name"`
}

type policyOptions struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type nodeOptions struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

// BuildPolicySettingRes 构建存储策略选项选择
func BuildPolicySettingRes(policies []model.Policy) Response {
	options := make([]policyOptions, 0, len(policies))
	for _, policy := range policies {
		options = append(options, policyOptions{
			Name: policy.Name,
			ID:   hashid.HashID(policy.ID, hashid.PolicyID),
		})
	}

	return Response{
		Data: options,
	}
}

// BuildMountedFolderRes 构建已挂载目录响应，list为当前用户可用存储策略ID
func BuildMountedFolderRes(folders []model.Folder, list []uint) []MountedFolders {
	res := make([]MountedFolders, 0, len(folders))
	for _, folder := range folders {
		single := MountedFolders{
			ID:         hashid.HashID(folder.ID, hashid.FolderID),
			Name:       folder.Name,
			PolicyName: "[Invalid Policy]",
		}
		if policy, err := model.GetPolicyByID(folder.PolicyID); err == nil && util.ContainsUint(list, policy.ID) {
			single.PolicyName = policy.Name
		}

		res = append(res, single)
	}

	return res
}

// BuildUserQuotaResponse 序列化用户存储配额概况响应
func BuildUserQuotaResponse(user *model.User, packs []model.StoragePack) Response {
	packSize := user.GetAvailablePackSize()
	res := quota{
		Base:  user.Group.MaxStorage,
		Pack:  packSize,
		Used:  user.Storage,
		Total: packSize + user.Group.MaxStorage,
		Packs: make([]storagePacks, 0, len(packs)),
	}
	for _, pack := range packs {
		res.Packs = append(res.Packs, storagePacks{
			Name:           pack.Name,
			Size:           pack.Size,
			ActivateDate:   *pack.ActiveTime,
			Expiration:     int(pack.ExpiredTime.Sub(*pack.ActiveTime).Seconds()),
			ExpirationDate: *pack.ExpiredTime,
		})
	}

	return Response{
		Data: res,
	}
}

// PackProduct 容量包商品
type PackProduct struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Size  uint64 `json:"size"`
	Time  int64  `json:"time"`
	Price int    `json:"price"`
	Score int    `json:"score"`
}

// GroupProducts 用户组商品
type GroupProducts struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	GroupID   uint     `json:"group_id"`
	Time      int64    `json:"time"`
	Price     int      `json:"price"`
	Score     int      `json:"score"`
	Des       []string `json:"des"`
	Highlight bool     `json:"highlight"`
}

// BuildProductResponse 构建增值服务商品响应
func BuildProductResponse(groups []GroupProducts, packs []PackProduct,
	wechat, alipay, payjs, custom bool, customName string, scorePrice int) Response {
	// 隐藏响应中的用户组ID
	for i := 0; i < len(groups); i++ {
		groups[i].GroupID = 0
	}
	return Response{
		Data: map[string]interface{}{
			"packs":       packs,
			"groups":      groups,
			"alipay":      alipay,
			"wechat":      wechat,
			"payjs":       payjs,
			"custom":      custom,
			"custom_name": customName,
			"score_price": scorePrice,
		},
	}
}

// BuildNodeOptionRes 构建可用节点列表响应
func BuildNodeOptionRes(nodes []*model.Node) Response {
	options := make([]nodeOptions, 0, len(nodes))
	for _, node := range nodes {
		options = append(options, nodeOptions{
			Name: node.Name,
			ID:   node.ID,
		})
	}

	return Response{
		Data: options,
	}
}
