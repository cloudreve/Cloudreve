package share

import (
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/service/explorer"
	"net/url"
)

type ListShareResponse struct {
	Shares     []explorer.Share             `json:"shares"`
	Pagination *inventory.PaginationResults `json:"pagination"`
}

func BuildListShareResponse(res *inventory.ListShareResult, hasher hashid.Encoder, base *url.URL, requester *ent.User, unlocked bool) *ListShareResponse {
	var infos []explorer.Share
	for _, share := range res.Shares {
		infos = append(infos, *explorer.BuildShare(share, base, hasher, requester, share.Edges.User, share.Edges.File.Name,
			types.FileType(share.Edges.File.Type), unlocked))
	}

	return &ListShareResponse{
		Shares:     infos,
		Pagination: res.PaginationResults,
	}
}
