package share

import (
	"net/url"

	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs/dbfs"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/service/explorer"
	"github.com/samber/lo"
)

type ListShareResponse struct {
	Shares     []explorer.Share             `json:"shares"`
	Pagination *inventory.PaginationResults `json:"pagination"`
}

func BuildListShareResponse(res *inventory.ListShareResult, hasher hashid.Encoder, base *url.URL, requester *ent.User, unlocked bool) *ListShareResponse {
	var infos []explorer.Share
	for _, share := range res.Shares {
		expired := inventory.IsValidShare(share) != nil
		shareName := share.Edges.File.Name
		if share.Edges.File.FileChildren == 0 && len(share.Edges.File.Edges.Metadata) >= 0 {
			// For files in trash bin, read the real name from metadata
			restoreUri, found := lo.Find(share.Edges.File.Edges.Metadata, func(m *ent.Metadata) bool {
				return m.Name == dbfs.MetadataRestoreUri
			})
			if found {
				uri, err := fs.NewUriFromString(restoreUri.Value)
				if err == nil {
					shareName = uri.Name()
				}
			}
		}

		infos = append(infos, *explorer.BuildShare(share, base, hasher, requester, share.Edges.User, shareName,
			types.FileType(share.Edges.File.Type), unlocked, expired))
	}

	return &ListShareResponse{
		Shares:     infos,
		Pagination: res.PaginationResults,
	}
}
