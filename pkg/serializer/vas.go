package serializer

import (
	model "github.com/HFO4/cloudreve/models"
)

type quota struct {
	Base  uint64         `json:"base"`
	Pack  uint64         `json:"pack"`
	Used  uint64         `json:"used"`
	Total uint64         `json:"total"`
	Packs []storagePacks `json:"packs"`
}

type storagePacks struct {
	Name           string `json:"name"`
	Size           uint64 `json:"size"`
	ActivateDate   string `json:"activate_date"`
	Expiration     int    `json:"expiration"`
	ExpirationDate string `json:"expiration_date"`
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
			ActivateDate:   pack.ActiveTime.Format("2006-01-02 15:04:05"),
			Expiration:     int(pack.ExpiredTime.Sub(*pack.ActiveTime).Seconds()),
			ExpirationDate: pack.ExpiredTime.Format("2006-01-02 15:04:05"),
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
func BuildProductResponse(groups []GroupProducts, packs []PackProduct, alipay, payjs bool) Response {
	// 隐藏响应中的用户组ID
	for i := 0; i < len(groups); i++ {
		groups[i].GroupID = 0
	}
	return Response{
		Data: map[string]interface{}{
			"packs":  packs,
			"groups": groups,
			"alipay": alipay,
			"payjs":  payjs,
		},
	}
}
